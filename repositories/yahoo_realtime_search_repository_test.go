package repositories

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetTimelineEntryDoesNotFollowRedirects(t *testing.T) {
	calls := 0
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			if req.URL.Scheme != "https" || req.URL.Host != "search.yahoo.co.jp" {
				t.Errorf("unexpected request URL: %s", req.URL)
			}

			return &http.Response{
				StatusCode: http.StatusFound,
				Status:     "302 Found",
				Header:     http.Header{"Location": []string{"http://127.0.0.1/internal"}},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		}),
	}

	_, _, err := (YahooRealtimeSearchRepository{HTTPClient: client}).GetTimelineEntry("test")
	if err == nil || !strings.Contains(err.Error(), "302") {
		t.Errorf("expected a redirect status error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected one request, got %d", calls)
	}
	if client.CheckRedirect != nil {
		t.Error("the caller's HTTP client was modified")
	}
}

func TestGetTimelineEntryReturnsErrQueryTooLong(t *testing.T) {
	body := `<html><body><script id="__NEXT_DATA__" type="application/json">` +
		`{"props":{"pageProps":{"pageData":{"query":"long","searchError":{"errorType":"longQuery"}}}}}` +
		`</script></body></html>`
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}),
	}

	_, _, err := (YahooRealtimeSearchRepository{HTTPClient: client}).GetTimelineEntry("long")
	if !errors.Is(err, ErrQueryTooLong) {
		t.Errorf("expected ErrQueryTooLong, got %v", err)
	}
}
