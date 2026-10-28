package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
	"github.com/kenchan0130/yahoo-realtime-search-feed/repositories"
)

const defaultIndexLimit = 10

const (
	formatRSS  = "rss"
	formatJSON = "json"
)

var feedContentTypes = map[string]string{
	formatRSS:  "application/xml; charset=utf-8",
	formatJSON: "application/feed+json; charset=utf-8",
}

type FeedController struct {
	YahooRealtimeSearchRepository *repositories.CachedTimelineRepository
}

// noCache reports whether the "nocache" parameter asks to skip the cache.
// "?nocache" without a value also skips it.
func noCache(query url.Values) bool {
	if !query.Has("nocache") {
		return false
	}

	v := strings.TrimSpace(query.Get("nocache"))
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	return err == nil && b
}

func (fc FeedController) Index(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	searchQuery := strings.TrimSpace(query.Get("q"))
	limitQuery := strings.TrimSpace(query.Get("limit"))
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		limit = defaultIndexLimit
	}
	format := strings.ToLower(strings.TrimSpace(query.Get("format")))
	if format == "" {
		format = formatRSS
	}
	contentType, ok := feedContentTypes[format]
	if !ok {
		http.Error(w, "Unsupported format. Use 'rss' or 'json'.", http.StatusBadRequest)
		return
	}

	getTimelineEntry := fc.YahooRealtimeSearchRepository.GetTimelineEntry
	if noCache(query) {
		getTimelineEntry = fc.YahooRealtimeSearchRepository.GetFreshTimelineEntry
	}
	entryList, res, err := getTimelineEntry(searchQuery)
	if errors.Is(err, repositories.ErrQueryTooLong) {
		http.Error(w, "The search query is too long. Please shorten it or reduce the number of terms.", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "Error. Please check server log.", http.StatusInternalServerError)
		return
	}

	requestURL := ""
	if res != nil {
		requestURL = res.Request.URL.String()
	}

	slicedEntryList := *entryList
	if len(slicedEntryList) > limit {
		slicedEntryList = slicedEntryList[:limit]
	}

	var body bytes.Buffer
	if err := fc.generateFeed(&body, format, slicedEntryList, searchQuery, requestURL); err != nil {
		log.Println(err)
		http.Error(w, "Error. Please check server log.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = body.WriteTo(w)
}

func (fc FeedController) generateFeed(buf *bytes.Buffer, format string, entryList []models.TimelineEntry, query string, requestURL string) error {
	// RSS readers show the title as the feed name, so it carries the query to tell feeds apart.
	withQuery := fmt.Sprintf(" with '%s'", query)
	if format == formatJSON {
		withQuery = ""
	}
	feed := &feeds.Feed{
		Title:       "Realtime Search Feed" + withQuery,
		Link:        &feeds.Link{Href: requestURL},
		Description: "This feed is Yahoo! Realtime Search" + withQuery,
		Image: &feeds.Image{
			Link:  requestURL,
			Url:   "https://abs.twimg.com/responsive-web/web/icon-default.3c3b2244.png", // From https://twitter.com/manifest.json
			Title: "Realtime Search Feed" + withQuery,
		},
	}

	feed.Items = make([]*feeds.Item, 0, len(entryList))
	for _, entry := range entryList {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       *entry.DisplayText,
			Link:        &feeds.Link{Href: *entry.URL},
			Description: *entry.URL,
			Created:     entry.CreatedAt.Time,
			Id:          *entry.ID,
			Author:      &feeds.Author{Name: *entry.ScreenName},
		})
	}

	if format == formatJSON {
		jsonFeed := (&feeds.JSON{Feed: feed}).JSONFeed()
		// JSON Feed requires "items" even when it is empty, but feeds.JSONFeed omits an empty one.
		items := jsonFeed.Items
		if items == nil {
			items = []*feeds.JSONItem{}
		}
		e := json.NewEncoder(buf)
		e.SetIndent("", "  ")
		if err := e.Encode(struct {
			*feeds.JSONFeed
			Items []*feeds.JSONItem `json:"items"`
		}{jsonFeed, items}); err != nil {
			return fmt.Errorf("json.Encoder#Encode(): %v", err)
		}
		return nil
	}

	if err := feed.WriteRss(buf); err != nil {
		return fmt.Errorf("feeds.Feed#WriteRss(): %v", err)
	}

	return nil
}
