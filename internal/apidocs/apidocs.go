// Package apidocs serves the OpenAPI contract and a viewer for it.
//
// The contract is how the frontend learns this API: there are no server-rendered
// pages any more, so the spec is the only description of what the backend does.
// Serving it from the backend keeps it honest — it ships in the same image as
// the code it describes and cannot drift into a stale copy somewhere else.
package apidocs

import (
	"net/http"
	"os"
)

// Handler serves the raw spec and a Swagger UI page for it.
type Handler struct {
	specPath string
}

// New reads the spec from specPath on each request rather than caching it, so a
// spec edited in development shows up on reload. The file is small.
func New(specPath string) *Handler {
	if specPath == "" {
		specPath = "api/openapi.yaml"
	}
	return &Handler{specPath: specPath}
}

// Spec serves the OpenAPI document itself.
func (h *Handler) Spec(w http.ResponseWriter, _ *http.Request) {
	raw, err := os.ReadFile(h.specPath)
	if err != nil {
		http.Error(w, "spec unavailable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(raw)
}

// UI serves a Swagger UI page pointed at the spec.
//
// It carries its own Content-Security-Policy: the site-wide one is written for a
// JSON API and blocks everything, while this one page legitimately loads the
// viewer from a CDN. Scoping the exception to this handler keeps the API's own
// responses locked down.
func (h *Handler) UI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy",
		"default-src 'none'; "+
			"script-src 'unsafe-inline' https://cdn.jsdelivr.net; "+
			"style-src 'unsafe-inline' https://cdn.jsdelivr.net; "+
			"img-src 'self' data:; font-src https://cdn.jsdelivr.net; "+
			"connect-src 'self'; base-uri 'none'")
	_, _ = w.Write([]byte(page))
}

const page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>TOT API</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
SwaggerUIBundle({
  url: "/openapi.yaml",
  dom_id: "#ui",
  deepLinking: true,
  // The API authenticates with a session cookie, so "Try it out" only works
  // from a browser that already signed in against this origin.
  withCredentials: true,
});
</script>
</body>
</html>
`
