package httpstat

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

type windowCounter struct {
	count       atomic.Int64
	windowStart atomic.Int64 // unix nano
}

func (w *windowCounter) inc() {
	w.count.Add(1)
}

func (w *windowCounter) load() int64 {
	return w.count.Load()
}

func (w *windowCounter) reset(now time.Time) int64 {
	w.windowStart.Store(now.UnixNano())
	return w.count.Swap(0)
}

type RouteStats struct {
	Method string
	Path   string
	total  atomic.Int64
	window windowCounter
}

func (rs *RouteStats) grandTotal() int64 {
	return rs.total.Load() + rs.window.load()
}

type Statistics struct {
	interval time.Duration

	GET    atomic.Int64
	POST   atomic.Int64
	PUT    atomic.Int64
	DELETE atomic.Int64
	Other  atomic.Int64
	Total  atomic.Int64

	mu     sync.RWMutex
	routes map[string]*RouteStats
}

func NewStatistics(interval time.Duration) *Statistics {
	return &Statistics{
		interval: interval,
		routes:   make(map[string]*RouteStats),
	}
}

func (s *Statistics) record(r *http.Request) {
	s.Total.Add(1)
	switch r.Method {
	case http.MethodGet:
		s.GET.Add(1)
	case http.MethodPost:
		s.POST.Add(1)
	case http.MethodPut:
		s.PUT.Add(1)
	case http.MethodDelete:
		s.DELETE.Add(1)
	default:
		s.Other.Add(1)
	}

	key, method, path := routeKey(r)

	s.mu.RLock()
	rs := s.routes[key]
	s.mu.RUnlock()

	if rs == nil {
		s.mu.Lock()
		if s.routes[key] == nil {
			rs = &RouteStats{Method: method, Path: path}
			rs.window.windowStart.Store(time.Now().UnixNano())
			s.routes[key] = rs
		} else {
			rs = s.routes[key]
		}
		s.mu.Unlock()
	}

	rs.window.inc()
}

func routeKey(r *http.Request) (key, method, path string) {
	method = r.Method
	path = r.URL.Path
	if route := mux.CurrentRoute(r); route != nil {
		if tmpl, err := route.GetPathTemplate(); err == nil {
			path = tmpl
		}
	}
	key = method + " " + path
	return
}

func Middleware(stats *Statistics) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stats.record(r)
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Statistics) StartWindowReset() {
	go func() {
		ticker := time.NewTicker(s.interval)
		for now := range ticker.C {
			s.mu.RLock()
			for _, rs := range s.routes {
				count := rs.window.reset(now)
				rs.total.Add(count)
			}
			s.mu.RUnlock()
		}
	}()
}

func Handler(stats *Statistics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		showRoutes := r.URL.Query().Get("routes") == "1"

		var b strings.Builder
		now := time.Now()
		fmt.Fprintf(&b, "=== HTTP Statistics (%s) ===\n\n", now.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&b, "%-10s %d\n", "Total:", stats.Total.Load())
		fmt.Fprintf(&b, "%-10s %d\n", "GET:", stats.GET.Load())
		fmt.Fprintf(&b, "%-10s %d\n", "POST:", stats.POST.Load())
		fmt.Fprintf(&b, "%-10s %d\n", "PUT:", stats.PUT.Load())
		fmt.Fprintf(&b, "%-10s %d\n", "DELETE:", stats.DELETE.Load())

		if showRoutes {
			stats.mu.RLock()
			routes := make([]*RouteStats, 0, len(stats.routes))
			for _, rs := range stats.routes {
				routes = append(routes, rs)
			}
			stats.mu.RUnlock()

			sort.Slice(routes, func(i, j int) bool {
				return routes[i].grandTotal() > routes[j].grandTotal()
			})

			windowHeader := fmt.Sprintf("WINDOW (%s)", stats.interval)
			fmt.Fprintf(&b, "\n%-6s %-45s %-12s %s\n", "METHOD", "PATH", windowHeader, "TOTAL")
			fmt.Fprintln(&b, strings.Repeat("-", 70))
			for _, rs := range routes {
				fmt.Fprintf(&b, "%-6s %-45s %-12d %d\n",
					rs.Method, rs.Path, rs.window.load(), rs.grandTotal())
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, b.String())
	}
}
