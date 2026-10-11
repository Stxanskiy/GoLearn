package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A cover in our own object store must come back from this service, not as a
// redirect to another origin: next/image refuses a host that is not in its
// remotePatterns, and that list is baked into the frontend at build time.
func TestCoverFromOurStoreIsStreamed(t *testing.T) {
	const png = "\x89PNG\r\n\x1a\nfake"
	store := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/covers/abc.png" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte(png))
	}))
	defer store.Close()

	users, c := storefront()
	// Point a course at the store and ask for its cover.
	for i := range c.modules {
		if c.modules[i].Slug == "linux" {
			c.modules[i].CoverImage = store.URL + "/covers/abc.png"
		}
	}
	h := newTestAPIWithMedia(t, users, c, store.URL)

	w := do(h, http.MethodGet, "/courses/linux/cover", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (a redirect means the browser leaves this origin)", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("content type = %q", got)
	}
	if w.Body.String() != png {
		t.Errorf("body = %q, want the image bytes", w.Body.String())
	}
}

// An address outside our store keeps the redirect. Fetching arbitrary URLs here
// would let anyone with the studio make this server request addresses only it
// can reach.
func TestCoverOutsideOurStoreRedirects(t *testing.T) {
	users, c := storefront()
	for i := range c.modules {
		if c.modules[i].Slug == "linux" {
			c.modules[i].CoverImage = "http://169.254.169.254/latest/meta-data/"
		}
	}
	h := newTestAPIWithMedia(t, users, c, "https://learn.example.com/golearn")

	w := do(h, http.MethodGet, "/courses/linux/cover", "")
	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302 — this must not be fetched by the server", w.Code)
	}
}

// A store that does not answer must not take the page down with it.
func TestCoverFallsBackWhenTheStoreIsDown(t *testing.T) {
	store := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer store.Close()

	users, c := storefront()
	for i := range c.modules {
		if c.modules[i].Slug == "linux" {
			c.modules[i].CoverImage = store.URL + "/covers/gone.png"
		}
	}
	h := newTestAPIWithMedia(t, users, c, store.URL)

	w := do(h, http.MethodGet, "/courses/linux/cover", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want the generated cover", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/svg+xml; charset=utf-8" {
		t.Errorf("content type = %q, want the generated SVG", got)
	}
}
