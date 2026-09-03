package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// loginLimiter is a tiny in-memory sliding-window limiter for login attempts,
// keyed by client IP — a first line against password brute-forcing. It is not a
// distributed limiter (per-instance), which is fine for the current single replica.
type loginLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

var loginRL = &loginLimiter{hits: map[string][]time.Time{}, max: 10, window: 10 * time.Minute}

// allow records an attempt and reports whether it is within the limit.
func (l *loginLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := time.Now().Add(-l.window)
	kept := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, time.Now())
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

// clientIP extracts the best-effort client IP, honouring X-Forwarded-For behind
// the ingress/proxy.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
