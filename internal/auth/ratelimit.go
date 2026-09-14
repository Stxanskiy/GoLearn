package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limiter is an in-memory per-instance sliding-window rate limiter keyed by client IP.
type Limiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

// NewLimiter allows max attempts per key within window.
func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{hits: map[string][]time.Time{}, max: max, window: window}
}

// LoginLimiter guards password login for both the HTML pages and the JSON API.
var LoginLimiter = NewLimiter(10, 10*time.Minute)

// RegisterLimiter guards self-registration.
var RegisterLimiter = NewLimiter(5, time.Hour)

// Allow records an attempt and reports whether it is within the limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := time.Now().Add(-l.window)
	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, time.Now())
	// opportunistic cleanup so the map does not grow unbounded
	if len(l.hits) > 10000 {
		for k, v := range l.hits {
			if len(v) == 0 || v[len(v)-1].Before(cut) {
				delete(l.hits, k)
			}
		}
	}
	return true
}

// Window is the limiter's time window (used for Retry-After).
func (l *Limiter) Window() time.Duration { return l.window }

// ClientIP extracts the best-effort client IP, honouring X-Forwarded-For behind
// the ingress/proxy.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
