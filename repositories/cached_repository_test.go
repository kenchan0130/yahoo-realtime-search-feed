package repositories

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
)

type timelineEntryGetterFunc func(string) (*[]models.TimelineEntry, *http.Response, error)

func (f timelineEntryGetterFunc) GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	return f(query)
}

func TestCachedTimelineRepository(t *testing.T) {
	t.Run("returns cached entries until they expire", func(t *testing.T) {
		var calls atomic.Int32
		repo := NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			calls.Add(1)
			return &[]models.TimelineEntry{}, nil, nil
		}), time.Minute, 10)
		now := time.Now()
		repo.now = func() time.Time { return now }

		for range 3 {
			if _, _, err := repo.GetTimelineEntry("q"); err != nil {
				t.Fatal(err)
			}
		}
		if got := calls.Load(); got != 1 {
			t.Errorf("expected one upstream call, got %d", got)
		}

		now = now.Add(time.Minute)
		if _, _, err := repo.GetTimelineEntry("q"); err != nil {
			t.Fatal(err)
		}
		if got := calls.Load(); got != 2 {
			t.Errorf("expected an upstream call after expiry, got %d calls", got)
		}
	})

	t.Run("GetFreshTimelineEntry skips the cache and refreshes it", func(t *testing.T) {
		var calls atomic.Int32
		repo := NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			n := calls.Add(1)
			return &[]models.TimelineEntry{{ID: pointer(strconv.Itoa(int(n)))}}, nil, nil
		}), time.Minute, 10)

		if _, _, err := repo.GetTimelineEntry("q"); err != nil {
			t.Fatal(err)
		}
		fresh, _, err := repo.GetFreshTimelineEntry("q")
		if err != nil {
			t.Fatal(err)
		}
		if got := *(*fresh)[0].ID; got != "2" {
			t.Errorf("expected fresh entries, got ID %s", got)
		}

		cached, _, err := repo.GetTimelineEntry("q")
		if err != nil {
			t.Fatal(err)
		}
		if got := *(*cached)[0].ID; got != "2" {
			t.Errorf("expected the cache to hold the fresh entries, got ID %s", got)
		}
		if got := calls.Load(); got != 2 {
			t.Errorf("expected two upstream calls, got %d", got)
		}
	})

	t.Run("does not cache errors", func(t *testing.T) {
		var calls atomic.Int32
		repo := NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			calls.Add(1)
			return nil, nil, errors.New("upstream error")
		}), time.Minute, 10)

		for range 2 {
			if _, _, err := repo.GetTimelineEntry("q"); err == nil {
				t.Fatal("expected an error")
			}
		}
		if got := calls.Load(); got != 2 {
			t.Errorf("expected two upstream calls, got %d", got)
		}
	})

	t.Run("merges concurrent requests with the same query", func(t *testing.T) {
		var calls atomic.Int32
		release := make(chan struct{})
		repo := NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			calls.Add(1)
			<-release
			return &[]models.TimelineEntry{}, nil, nil
		}), 0, 10)

		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				if _, _, err := repo.GetTimelineEntry("q"); err != nil {
					t.Error(err)
				}
			})
		}
		time.Sleep(50 * time.Millisecond)
		close(release)
		wg.Wait()

		if got := calls.Load(); got != 1 {
			t.Errorf("expected one upstream call, got %d", got)
		}
	})

	t.Run("keeps at most maxEntries queries", func(t *testing.T) {
		repo := NewCachedTimelineRepository(timelineEntryGetterFunc(func(string) (*[]models.TimelineEntry, *http.Response, error) {
			return &[]models.TimelineEntry{}, nil, nil
		}), time.Minute, 2)

		for _, q := range []string{"a", "b", "c", "d"} {
			if _, _, err := repo.GetTimelineEntry(q); err != nil {
				t.Fatal(err)
			}
		}
		if got := len(repo.cache); got != 2 {
			t.Errorf("expected 2 cached queries, got %d", got)
		}
	})
}

func pointer[T any](v T) *T {
	return &v
}
