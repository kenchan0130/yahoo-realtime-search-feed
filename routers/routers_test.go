package routers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestInitRouter(t *testing.T) {
	t.Run("GET /health returns 'ok' with status 200", func(t *testing.T) {
		router := Init()

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, "ok", w.Body.String())
		assert.Equal(t, 200, w.Code)
	})

	t.Run("GET /feed returns rss with status 200", func(t *testing.T) {
		router := Init()

		w := httptest.NewRecorder()

		q := "Twitter"
		limit := "1"

		queryParameters := url.Values{
			"q":     {q},
			"limit": {limit},
		}

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/feed?%s", queryParameters.Encode()), nil)
		router.ServeHTTP(w, req)

		fp := gofeed.NewParser()
		feed, err := fp.Parse(bytes.NewReader(w.Body.Bytes()))

		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, fmt.Sprintf("Realtime Search Feed with '%s'", q), feed.Title)
		assert.Equal(t, limit, fmt.Sprint(len(feed.Items)))
	})

	t.Run("GET /feed returns empty RSS when no match", func(t *testing.T) {
		router := Init()

		w := httptest.NewRecorder()

		randomString, _ := uuid.NewRandom()
		q := randomString.String()
		limit := "1"

		queryParameters := url.Values{
			"q":     {q},
			"limit": {limit},
		}

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/feed?%s", queryParameters.Encode()), nil)
		router.ServeHTTP(w, req)

		fp := gofeed.NewParser()
		feed, err := fp.Parse(bytes.NewReader(w.Body.Bytes()))

		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, fmt.Sprintf("Realtime Search Feed with '%s'", q), feed.Title)
		assert.Equal(t, 0, len(feed.Items))
	})
}
