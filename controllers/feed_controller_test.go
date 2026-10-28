package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
	"github.com/kenchan0130/yahoo-realtime-search-feed/repositories"
)

func TestNoCache(t *testing.T) {
	for raw, want := range map[string]bool{
		"":             false,
		"q=a":          false,
		"nocache":      true,
		"nocache=":     true,
		"nocache=1":    true,
		"nocache=true": true,
		"nocache=0":    false,
		"nocache=no":   false,
	} {
		query, err := url.ParseQuery(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := noCache(query); got != want {
			t.Errorf("noCache(%q) = %v, want %v", raw, got, want)
		}
	}
}

type timelineEntryGetterFunc func(string) (*[]models.TimelineEntry, *http.Response, error)

func (f timelineEntryGetterFunc) GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	return f(query)
}

func TestIndexReturnsBadRequestForTooLongQuery(t *testing.T) {
	fc := FeedController{
		YahooRealtimeSearchRepository: repositories.NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			return nil, nil, repositories.ErrQueryTooLong
		}), time.Minute, 10),
	}

	w := httptest.NewRecorder()
	fc.Index(w, httptest.NewRequest(http.MethodGet, "/feed?q=long", nil))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("expected a text response, got %q", got)
	}
	if got := w.Body.String(); !strings.Contains(got, "too long") {
		t.Errorf("expected the reason in the body, got %q", got)
	}
}

func TestIndexFormat(t *testing.T) {
	entry := models.TimelineEntry{}
	if err := json.Unmarshal([]byte(`{"id":"1","url":"https://x.com/a/status/1","displayText":"hello","createdAt":1700000000,"screenName":"a"}`), &entry); err != nil {
		t.Fatal(err)
	}
	var calls int
	fc := FeedController{
		YahooRealtimeSearchRepository: repositories.NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			calls++
			return &[]models.TimelineEntry{entry}, nil, nil
		}), 0, 10),
	}

	for _, tt := range []struct {
		format      string
		status      int
		contentType string
	}{
		{"", http.StatusOK, "application/xml"},
		{"rss", http.StatusOK, "application/xml"},
		{"json", http.StatusOK, "application/feed+json"},
		{"JSON", http.StatusOK, "application/feed+json"},
		{"atom", http.StatusBadRequest, "text/plain"},
	} {
		w := httptest.NewRecorder()
		fc.Index(w, httptest.NewRequest(http.MethodGet, "/feed?q=a&format="+tt.format, nil))

		if w.Code != tt.status {
			t.Errorf("format=%q: expected status %d, got %d", tt.format, tt.status, w.Code)
		}
		if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, tt.contentType) {
			t.Errorf("format=%q: expected Content-Type %q, got %q", tt.format, tt.contentType, got)
		}
		if tt.contentType == "application/feed+json" {
			var feed struct {
				Title string `json:"title"`
				Items []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"items"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &feed); err != nil {
				t.Fatalf("format=%q: invalid JSON: %v", tt.format, err)
			}
			if feed.Title != "Realtime Search Feed" || len(feed.Items) != 1 || feed.Items[0].ID != "1" || feed.Items[0].Title != "hello" {
				t.Errorf("format=%q: unexpected feed %+v", tt.format, feed)
			}
		}
	}
	t.Run("empty items", func(t *testing.T) {
		fc := FeedController{
			YahooRealtimeSearchRepository: repositories.NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
				return &[]models.TimelineEntry{}, nil, nil
			}), 0, 10),
		}
		w := httptest.NewRecorder()
		fc.Index(w, httptest.NewRequest(http.MethodGet, "/feed?q=a&format=json", nil))

		var feed map[string]json.RawMessage
		if err := json.Unmarshal(w.Body.Bytes(), &feed); err != nil {
			t.Fatal(err)
		}
		if got := string(feed["items"]); got != "[]" {
			t.Errorf("expected empty items, got %q", got)
		}
	})

	if calls != 4 {
		t.Errorf("expected an unsupported format to be rejected before searching, got %d searches", calls)
	}
}
