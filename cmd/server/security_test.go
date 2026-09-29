package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The service-wide policy forbids framing, which is right for a JSON API and
// wrong for the two handlers that return a page: /docs and the lab preview.
// Both override it by setting the header themselves, which only works because
// the middleware writes its defaults before the handler runs. If that order
// ever flips, the overrides lose silently — the handler still looks correct and
// the preview still refuses to frame.
func TestSecurityHeadersCanBeOverriddenByHandlers(t *testing.T) {
	const own = "default-src * data: blob: 'unsafe-inline' 'unsafe-eval'; frame-ancestors 'self'"
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", own)
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/1/lab/preview/80/", nil))

	if got := w.Header().Get("Content-Security-Policy"); got != own {
		t.Errorf("handler policy was overwritten by the middleware: %q", got)
	}
	// The headers a page has no reason to override must still be applied.
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options is %q", got)
	}
}

// Everything else is a JSON API and must stay unframeable.
func TestSecurityHeadersDefaultForbidsFraming(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil))

	if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Errorf("the API can be framed, CSP is %q", csp)
	}
}
