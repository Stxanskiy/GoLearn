// Package obs holds the service's observability surface: Prometheus metrics and
// the HTTP middleware that feeds them.
//
// The point is to answer "what is slow / what is failing / how loaded are we"
// without reading logs. Every incident so far has been diagnosed by hand —
// tailing journald, counting processes — because there was nothing to look at.
package obs

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry is this service's own registry rather than the default one, so a
// dependency that registers metrics on init cannot collide with ours.
var Registry = prometheus.NewRegistry()

var (
	requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "golearn_http_requests_total",
		Help: "HTTP requests by route pattern, method and status class.",
	}, []string{"route", "method", "status"})

	duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "golearn_http_request_duration_seconds",
		Help: "HTTP request duration by route pattern.",
		// Terminal and upload routes live in seconds, plain JSON in milliseconds,
		// so the buckets have to span both without drowning either.
		Buckets: []float64{0.005, 0.025, 0.1, 0.25, 1, 2.5, 10, 30},
	}, []string{"route", "method"})

	inFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "golearn_http_requests_in_flight",
		Help: "HTTP requests currently being served.",
	})
)

func init() {
	Registry.MustRegister(requests, duration, inFlight)
	Registry.MustRegister(collectors.NewGoCollector())
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// Gauge registers a gauge fed by fn at scrape time. Used for numbers that are
// owned elsewhere — pool sizes, running micro-VMs — so those packages do not
// have to import Prometheus.
func Gauge(name, help string, fn func() float64) {
	Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{Name: name, Help: help}, fn))
}

// Counter registers a labelled counter and hands back a function that bumps it,
// so callers can record an event without importing Prometheus themselves. Keep
// the label values to a small fixed set — one series is minted per combination.
func Counter(name, help string, labels ...string) func(...string) {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, labels)
	Registry.MustRegister(c)
	return func(values ...string) { c.WithLabelValues(values...).Inc() }
}

// Handler serves the metrics in Prometheus text format.
func Handler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{Registry: Registry})
}

// Middleware records every request. The route label is chi's pattern, never the
// raw path: labelling by path would mint a new time series per lesson id and
// blow up the registry.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inFlight.Inc()
		defer inFlight.Dec()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		requests.WithLabelValues(route, r.Method, statusClass(ww.Status())).Inc()
		duration.WithLabelValues(route, r.Method).Observe(time.Since(start).Seconds())
	})
}

// statusClass keeps the cardinality at five values instead of one per code; the
// exact status stays in the request log.
func statusClass(code int) string {
	switch {
	case code == 0:
		return "none"
	case code < 200:
		return "1xx"
	case code < 300:
		return "2xx"
	case code < 400:
		return "3xx"
	case code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}
