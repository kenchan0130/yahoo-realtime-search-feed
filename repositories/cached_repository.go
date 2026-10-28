package repositories

import (
	"net/http"
	"sync"
	"time"

	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
	"golang.org/x/sync/singleflight"
)

type TimelineEntryGetter interface {
	GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error)
}

type cachedTimeline struct {
	entryList *[]models.TimelineEntry
	res       *http.Response
	expiresAt time.Time
}

// CachedTimelineRepository keeps results of the wrapped repository for a while,
// and merges concurrent requests with the same query into one upstream request.
// Callers share the returned entries, so they must not modify them.
type CachedTimelineRepository struct {
	repository TimelineEntryGetter
	ttl        time.Duration
	maxEntries int
	now        func() time.Time

	group singleflight.Group
	mu    sync.RWMutex
	cache map[string]cachedTimeline
}

func NewCachedTimelineRepository(repository TimelineEntryGetter, ttl time.Duration, maxEntries int) *CachedTimelineRepository {
	return &CachedTimelineRepository{
		repository: repository,
		ttl:        ttl,
		maxEntries: maxEntries,
		now:        time.Now,
		cache:      map[string]cachedTimeline{},
	}
}

func (c *CachedTimelineRepository) GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	if v, ok := c.get(query); ok {
		return v.entryList, v.res, nil
	}

	v, err, _ := c.group.Do(query, func() (any, error) {
		// Another request may have filled the cache while this one was waiting.
		if v, ok := c.get(query); ok {
			return v, nil
		}

		entryList, res, err := c.repository.GetTimelineEntry(query)
		if err != nil {
			return nil, err //nolint:wrapcheck // The wrapped repository already adds context.
		}

		v := cachedTimeline{entryList: entryList, res: res, expiresAt: c.now().Add(c.ttl)}
		c.set(query, v)
		return v, nil
	})
	if err != nil {
		return nil, nil, err //nolint:wrapcheck // The wrapped repository already adds context.
	}

	cached := v.(cachedTimeline)
	return cached.entryList, cached.res, nil
}

// GetFreshTimelineEntry skips the cache, and stores the fetched result for later requests.
func (c *CachedTimelineRepository) GetFreshTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	// A separate key keeps this from joining a request that may answer from the cache.
	v, err, _ := c.group.Do("fresh\x00"+query, func() (any, error) {
		entryList, res, err := c.repository.GetTimelineEntry(query)
		if err != nil {
			return nil, err //nolint:wrapcheck // The wrapped repository already adds context.
		}

		v := cachedTimeline{entryList: entryList, res: res, expiresAt: c.now().Add(c.ttl)}
		c.set(query, v)
		return v, nil
	})
	if err != nil {
		return nil, nil, err //nolint:wrapcheck // The wrapped repository already adds context.
	}

	cached := v.(cachedTimeline)
	return cached.entryList, cached.res, nil
}

func (c *CachedTimelineRepository) get(query string) (cachedTimeline, bool) {
	if c.ttl <= 0 {
		return cachedTimeline{}, false
	}

	c.mu.RLock()
	v, ok := c.cache[query]
	c.mu.RUnlock()
	if !ok || !c.now().Before(v.expiresAt) {
		return cachedTimeline{}, false
	}

	return v, true
}

func (c *CachedTimelineRepository) set(query string, v cachedTimeline) {
	if c.ttl <= 0 || c.maxEntries <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[query]; !ok && len(c.cache) >= c.maxEntries {
		now := c.now()
		for k, cached := range c.cache {
			if !now.Before(cached.expiresAt) {
				delete(c.cache, k)
			}
		}
		// When everything is still fresh, drop an arbitrary entry; map iteration order is random.
		for k := range c.cache {
			if len(c.cache) < c.maxEntries {
				break
			}
			delete(c.cache, k)
		}
	}
	c.cache[query] = v
}
