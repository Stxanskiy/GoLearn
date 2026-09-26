package obs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMiddlewareLabelsByRoutePatternNotPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Middleware)
	r.Get("/lessons/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
	r.Get("/boom", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(500) })

	for _, path := range []string{"/lessons/1", "/lessons/2", "/lessons/3"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/boom", nil))

	body := scrape(t)
	// Three different ids must collapse into one series, or every lesson mints its
	// own and the registry grows without bound.
	if !strings.Contains(body, `golearn_http_requests_total{method="GET",route="/lessons/{id}",status="2xx"} 3`) {
		t.Error("requests are not collapsed under the route pattern")
	}
	if strings.Contains(body, `route="/lessons/1"`) {
		t.Error("raw path leaked into a label")
	}
	if !strings.Contains(body, `route="/boom",status="5xx"} 1`) {
		t.Error("server error not counted as 5xx")
	}
}

func TestGaugeReportsAtScrapeTime(t *testing.T) {
	n := 1.0
	Gauge("golearn_test_gauge", "help", func() float64 { return n })
	if !strings.Contains(scrape(t), "golearn_test_gauge 1") {
		t.Fatal("gauge missing")
	}
	n = 42
	if !strings.Contains(scrape(t), "golearn_test_gauge 42") {
		t.Error("gauge was sampled once instead of at scrape time")
	}
}

func scrape(t *testing.T) string {
	t.Helper()
	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	if w.Code != 200 {
		t.Fatalf("metrics returned %d", w.Code)
	}
	return w.Body.String()
}
