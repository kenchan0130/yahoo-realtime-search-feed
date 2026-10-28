package repositories

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	json "github.com/goccy/go-json"
	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
)

var (
	// A longer marker lets bytes.Index skip ahead faster, and the shorter one covers other attribute orders.
	nextDataScriptMarker = []byte(`<script id="__NEXT_DATA__"`)
	nextDataMarker       = []byte(`id="__NEXT_DATA__"`)
	nomatchMarker        = []byte(`id="nomatch"`)
	scriptEndTag         = []byte(`</script>`)

	// The response HTML is several hundred KB, so read buffers are reused across requests.
	bufferPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

const maxPooledBufferSize = 4 << 20

// ErrQueryTooLong is returned when Yahoo! Realtime Search rejects the query as too long.
var ErrQueryTooLong = errors.New("the search query is too long")

type YahooRealtimeSearchRepository struct {
	HTTPClient *http.Client
}

func (t YahooRealtimeSearchRepository) GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	if query == "" {
		return &[]models.TimelineEntry{}, nil, nil
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "search.yahoo.co.jp",
		Path:     "/realtime/search",
		RawQuery: url.Values{"p": {query}, "ei": {"UTF-8"}}.Encode(),
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, nil, fmt.Errorf("http.NewRequestWithContext(): %v", err)
	}

	// The request target is fixed, and redirects must not send it to another host.
	client := *t.HTTPClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http.Client#Do(): %v", err)
	}
	// Closing the body lets the transport reuse the connection for the next request.
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPooledBufferSize {
			bufferPool.Put(buf)
		}
	}()
	if res.ContentLength > 0 && res.ContentLength <= maxPooledBufferSize {
		buf.Grow(int(res.ContentLength))
	}
	if _, err := io.Copy(buf, res.Body); err != nil {
		return nil, res, fmt.Errorf("io.Copy(): %v", err)
	}
	// json.Unmarshal copies its input, so nothing returned refers to buf.
	body := buf.Bytes()

	// When there are no hits in the search results, the JSON in data may not exist, so empty is returned.
	// This is checked only on failure paths, because scanning the whole HTML for it is not free.
	isNoMatch := func() bool { return bytes.Contains(body, nomatchMarker) }

	dataBytes := extractNextData(body)
	if len(dataBytes) == 0 {
		if isNoMatch() {
			return &[]models.TimelineEntry{}, res, nil
		}
		return nil, res, fmt.Errorf("'#__NEXT_DATA__' DOM is not found, please check %s response", u.String())
	}

	var data struct {
		Props *models.Props `json:"props,omitempty"`
	}
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, res, fmt.Errorf("json.Unmarshal(): %v", err)
	}

	if data.Props != nil && data.Props.PageProps != nil && data.Props.PageProps.PageData != nil {
		if searchError := data.Props.PageProps.PageData.SearchError; searchError != nil && searchError.ErrorType != nil && *searchError.ErrorType == "longQuery" {
			return nil, res, ErrQueryTooLong
		}
	}
	if data.Props == nil || data.Props.PageProps == nil || data.Props.PageProps.PageData == nil || data.Props.PageProps.PageData.Timeline == nil {
		if isNoMatch() {
			return &[]models.TimelineEntry{}, res, nil
		}
	}
	if data.Props == nil {
		return nil, res, fmt.Errorf("the data of Props is not found, please check %s response", u.String())
	}
	if data.Props.PageProps == nil {
		return nil, res, fmt.Errorf("the data of Props.PageProps is not found, please check %s response", u.String())
	}
	if data.Props.PageProps.PageData == nil {
		return nil, res, fmt.Errorf("the data of Props.PageProps.PageData is not found, please check %s response", u.String())
	}
	if data.Props.PageProps.PageData.Timeline == nil {
		return nil, res, fmt.Errorf("the data of Props.PageProps.PageData.Timeline is not found, please check %s response", u.String())
	}

	entryList := data.Props.PageProps.PageData.Timeline.Entry
	if entryList == nil {
		return &[]models.TimelineEntry{}, res, nil
	}

	// Normalize display text about enclosure strings
	for i := range *entryList {
		item := &(*entryList)[i]
		item.DisplayText = normalizeEnclosure(item.DisplayText)
		item.DisplayTextBody = normalizeEnclosure(item.DisplayTextBody)
		item.DisplayTextFragments = normalizeEnclosure(item.DisplayTextFragments)
	}

	return entryList, res, nil
}

// extractNextData returns the content of <script id="__NEXT_DATA__">...</script> without building a DOM.
func extractNextData(body []byte) []byte {
	if i := bytes.Index(body, nextDataScriptMarker); i >= 0 {
		body = body[i+len(nextDataScriptMarker):]
	} else if i := bytes.Index(body, nextDataMarker); i >= 0 {
		body = body[i+len(nextDataMarker):]
	} else {
		return nil
	}

	start := bytes.IndexByte(body, '>')
	if start < 0 {
		return nil
	}
	body = body[start+1:]

	end := bytes.Index(body, scriptEndTag)
	if end < 0 {
		return nil
	}

	return bytes.TrimSpace(body[:end])
}

const (
	enclosureStart = "\tSTART\t"
	enclosureEnd   = "\tEND\t"
)

// normalizeEnclosure replaces "\tSTART\t<word>\tEND\t" with "<word>".
// It behaves like replacing `\tSTART\t(.*?)\tEND\t` with its non-empty group, without the cost of regexp.
func normalizeEnclosure(s *string) *string {
	if s == nil || !strings.Contains(*s, enclosureStart) {
		return s
	}

	rest := *s
	var b strings.Builder
	b.Grow(len(rest))
	for {
		i := strings.Index(rest, enclosureStart)
		if i < 0 {
			break
		}
		inner := rest[i+len(enclosureStart):]
		j := strings.Index(inner, enclosureEnd)
		// "." in the original pattern does not match a newline, so the enclosure must not span lines.
		if j < 0 || strings.IndexByte(inner[:j], '\n') >= 0 {
			b.WriteString(rest[:i+1])
			rest = rest[i+1:]
			continue
		}
		b.WriteString(rest[:i])
		if j == 0 {
			b.WriteString(rest[i : i+len(enclosureStart)+len(enclosureEnd)])
		} else {
			b.WriteString(inner[:j])
		}
		rest = inner[j+len(enclosureEnd):]
	}
	b.WriteString(rest)

	normalized := b.String()
	return &normalized
}
