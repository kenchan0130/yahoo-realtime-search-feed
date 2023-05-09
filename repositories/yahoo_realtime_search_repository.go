package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
	"github.com/kenchan0130/yahoo-realtime-search-feed/utils"
	"github.com/samber/lo"
)

type YahooRealtimeSearchRepository struct {
	HTTPClient *http.Client
}

func (t YahooRealtimeSearchRepository) GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	if query == "" {
		return &[]models.TimelineEntry{}, nil, nil
	}

	u, err := url.Parse("https://search.yahoo.co.jp/realtime/search")
	if err != nil {
		return nil, nil, fmt.Errorf("url.Parse(): %v", err)
	}

	queryParameter := url.Values{
		"p":  {query},
		"ei": {"UTF-8"},
	}

	u.RawQuery = queryParameter.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, nil, fmt.Errorf("http.NewRequestWithContext(): %v", err)
	}

	res, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http.Client#Do(): %v", err)
	}
	defer req.Body.Close()

	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, res, fmt.Errorf("goquery.NewDocumentFromReader(): %v", err)
	}

	// When there are no hits in the search results, the JSON in data does not exist, so empty is returned.
	if doc.Find("#nomatch").Size() > 0 {
		return &[]models.TimelineEntry{}, res, nil
	}

	dataStr := ""
	doc.Find("#__NEXT_DATA__").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		dataStr = selection.Text()
		return false
	})
	if dataStr == "" {
		return nil, res, fmt.Errorf("'#__NEXT_DATA__' DOM is not found, please check %s response", u.String())
	}

	var data struct {
		Props *models.Props `json:"props,omitempty"`
	}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, res, fmt.Errorf("json.Unmarshal(): %v", err)
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

	enclosureRegexp, err := regexp.Compile(`\tSTART\t(.*?)\tEND\t`)
	if err != nil {
		return nil, nil, fmt.Errorf("regexp.Compile(): %v", err)
	}

	cb := func(s string) string {
		matches := enclosureRegexp.FindStringSubmatch(s)
		if len(matches) == 2 {
			if word := matches[1]; word != "" {
				return word
			} else {
				return s
			}
		}
		return s
	}

	// Normalize display text about enclosure strings
	data.Props.PageProps.PageData.Timeline.Entry = utils.Pointer(lo.Map(*data.Props.PageProps.PageData.Timeline.Entry, func(item models.TimelineEntry, _ int) models.TimelineEntry {
		item.DisplayText = utils.Pointer(enclosureRegexp.ReplaceAllStringFunc(*item.DisplayText, cb))
		item.DisplayTextBody = utils.Pointer(enclosureRegexp.ReplaceAllStringFunc(*item.DisplayTextBody, cb))
		item.DisplayTextFragments = utils.Pointer(enclosureRegexp.ReplaceAllStringFunc(*item.DisplayTextFragments, cb))

		return item
	}))

	return data.Props.PageProps.PageData.Timeline.Entry, res, nil
}
