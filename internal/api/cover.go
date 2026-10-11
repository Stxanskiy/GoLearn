package api

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/backendraz/golearn/internal/catalog"
)

// coverClient fetches covers out of our own object store. Short timeouts: a
// cover is a few hundred kilobytes from a service one hop away, and a page of
// the catalogue waits on twelve of them.
var coverClient = &http.Client{Timeout: 10 * time.Second}

// serveCover answers with a course or specialization cover.
//
// A cover in our own object store is streamed through this service rather than
// redirected to. The redirect was correct and still broke every cover that
// moved to MinIO: the browser ended up on a different origin, and next/image
// refuses a host that is not in its remotePatterns — a list Next bakes at build
// time from S3_PUBLIC_URL, which the frontend's build does not have. The image
// then fails with nothing in any log, because from the server's side it was a
// perfectly good 302.
//
// Streaming keeps the cover same-origin, so it works with no second setting to
// keep in step across two deployments and a build.
//
// Only our own store is streamed. An author may paste any http(s) address into
// "обложка по ссылке", and fetching those here would turn this endpoint into a
// way to make the server request arbitrary addresses, including ones only it
// can reach. Those keep the redirect: the browser fetches them itself, which is
// what pasting an external link asks for.
func (a *API) serveCover(w http.ResponseWriter, r *http.Request, image, fallbackSVG string) {
	if !a.ours(image) {
		catalog.ServeCover(w, r, image, fallbackSVG)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, image, nil)
	if err != nil {
		catalog.ServeCover(w, r, "", fallbackSVG)
		return
	}
	resp, err := coverClient.Do(req)
	if err != nil {
		// The store is unreachable or slow. A course with no picture is a far
		// better page than a broken one, and the generated cover is what every
		// course without an upload already shows.
		a.log.Warn("cover: object store unreachable", "url", image, "error", err)
		catalog.ServeCover(w, r, "", fallbackSVG)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(ct, "image/") {
		a.log.Warn("cover: object store answered oddly", "url", image, "status", resp.StatusCode, "type", ct)
		catalog.ServeCover(w, r, "", fallbackSVG)
		return
	}

	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// The URL carries a hash of the image, so a cover that changes gets a new
	// address and this can be cached hard.
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if n := resp.Header.Get("Content-Length"); n != "" {
		w.Header().Set("Content-Length", n)
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		a.log.Warn("cover: copy interrupted", "url", image, "error", err)
	}
}

// ours reports whether the address is in the object store this service writes
// to. Compared against the configured public base and required to be a path
// under it, so a host that merely starts with the same text does not pass.
func (a *API) ours(image string) bool {
	base := strings.TrimRight(a.cfg.MediaURL, "/")
	return base != "" && strings.HasPrefix(image, base+"/")
}
