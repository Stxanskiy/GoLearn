package api

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
)

// catalogFilter narrows the catalogue by the query parameters of the request.
type catalogFilter struct {
	spec       string
	difficulty string
	status     string
	query      string
}

// Length limits of the free-text filter parameters, as the contract declares them.
const (
	catalogSpecMax  = 80
	catalogQueryMax = 100
)

// catalogFilterFrom reads the filter, writing the error response itself when a value is unknown.
func catalogFilterFrom(w http.ResponseWriter, r *http.Request) (catalogFilter, bool) {
	q := r.URL.Query()
	f := catalogFilter{
		spec:       strings.TrimSpace(q.Get("spec")),
		difficulty: strings.TrimSpace(q.Get("difficulty")),
		status:     strings.TrimSpace(q.Get("status")),
		query:      strings.ToLower(strings.TrimSpace(q.Get("q"))),
	}
	fields := map[string]string{}
	if f.difficulty != "" && !apigen.Difficulty(f.difficulty).Valid() {
		fields["difficulty"] = fieldInvalidFormat
	}
	if f.status != "" && !apigen.ProgressStatus(f.status).Valid() {
		fields["status"] = fieldInvalidFormat
	}
	if utf8.RuneCountInString(f.spec) > catalogSpecMax {
		fields["spec"] = fieldTooLong
	}
	if utf8.RuneCountInString(f.query) > catalogQueryMax {
		fields["q"] = fieldTooLong
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return catalogFilter{}, false
	}
	return f, true
}

// matches reports whether a course survives every set filter.
func (f catalogFilter) matches(c catalog.Course) bool {
	if f.spec != "" && c.Spec != f.spec {
		return false
	}
	if f.difficulty != "" && c.Module.Difficulty != f.difficulty {
		return false
	}
	if f.status != "" && c.Status != f.status {
		return false
	}
	return f.query == "" || matchesQuery(c, f.query)
}

// matchesQuery looks the needle up in the fields the catalogue search covers.
func matchesQuery(c catalog.Course, needle string) bool {
	haystack := []string{c.Module.Title, c.Module.Description, c.Category}
	haystack = append(haystack, c.Module.Tags...)
	for _, field := range haystack {
		if strings.Contains(strings.ToLower(field), needle) {
			return true
		}
	}
	return false
}

// keep returns the courses that survive the filter; the full set stays available for the counters.
func (f catalogFilter) keep(courses []catalog.Course) []catalog.Course {
	shown := make([]catalog.Course, 0, len(courses))
	for _, c := range courses {
		if f.matches(c) {
			shown = append(shown, c)
		}
	}
	return shown
}
