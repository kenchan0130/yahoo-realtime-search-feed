package routers

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kenchan0130/yahoo-realtime-search-feed/controllers"
	"github.com/kenchan0130/yahoo-realtime-search-feed/repositories"
)

const (
	defaultCacheTTL     = 5 * time.Minute
	maxCachedQueryCount = 1000
)

// cacheTTL reads CACHE_TTL (e.g. "30s", "1m"). "0" disables the cache.
func cacheTTL() time.Duration {
	v := os.Getenv("CACHE_TTL")
	if v == "" {
		return defaultCacheTTL
	}
	if v == "0" {
		return 0
	}

	ttl, err := time.ParseDuration(v)
	if err != nil || ttl < 0 {
		log.Printf("invalid CACHE_TTL %q, using %v", v, defaultCacheTTL) //nolint:gosec // The value is quoted with %q.
		return defaultCacheTTL
	}

	return ttl
}

func Init() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/health", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	// Every request goes to the same host, so keep enough idle connections to it to skip TCP and TLS handshakes.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	yahooRealtimeSearchRepository := repositories.YahooRealtimeSearchRepository{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
	feedCtrl := controllers.FeedController{
		YahooRealtimeSearchRepository: repositories.NewCachedTimelineRepository(yahooRealtimeSearchRepository, cacheTTL(), maxCachedQueryCount),
	}
	// GET patterns also match HEAD requests, and net/http drops the body for them.
	mux.HandleFunc("GET /feed", feedCtrl.Index)

	return recovery(logger(mux))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		// Request values are quoted with %q, so control characters cannot forge log lines.
		log.Printf("| %3d | %13v | %q | %q %q", rec.status, time.Since(start), r.RemoteAddr, r.Method, r.URL.Path) //nolint:gosec
	})
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if err == http.ErrAbortHandler {
					panic(err)
				}
				log.Printf("panic recovered: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
