package apidocs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpecServesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := os.WriteFile(path, []byte("openapi: 3.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	New(path).Spec(rec, httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/yaml") {
		t.Errorf("content type %q", ct)
	}
	if rec.Body.String() != "openapi: 3.1.0\n" {
		t.Errorf("body %q", rec.Body.String())
	}
}

func TestSpecMissingFileIs404(t *testing.T) {
	rec := httptest.NewRecorder()
	New(filepath.Join(t.TempDir(), "absent.yaml")).Spec(rec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

// The viewer is served at /docs and at /api/v1/docs, so it must find the spec
// next to itself; an absolute URL would send the /api/v1 copy to a path the
// public site's proxy does not pass through.
func TestUILoadsTheSpecRelatively(t *testing.T) {
	rec := httptest.NewRecorder()
	New("").UI(rec, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))

	body := rec.Body.String()
	if !strings.Contains(body, `url: "openapi.yaml"`) {
		t.Error("viewer does not load the spec by a relative URL")
	}
	if csp := rec.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "https://cdn.jsdelivr.net") {
		t.Errorf("viewer CSP does not allow its CDN: %q", csp)
	}
}
