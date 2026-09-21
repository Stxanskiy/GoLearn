package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/model"
)

func filterCourses() []catalog.Course {
	return []catalog.Course{
		{Module: model.Module{Slug: "linux", Title: "Linux: Старт", Difficulty: "beginner", Tags: []string{"bash"}},
			Spec: "devops", Category: "Linux", Status: catalog.StatusCompleted},
		{Module: model.Module{Slug: "docker", Title: "Docker", Description: "Контейнеры", Difficulty: "intermediate"},
			Spec: "devops", Category: "Docker", Status: catalog.StatusInProgress},
		{Module: model.Module{Slug: "sql", Title: "SQL", Difficulty: "beginner"},
			Spec: "database", Category: "Database", Status: catalog.StatusNotStarted},
	}
}

func filterFrom(t *testing.T, query string) catalogFilter {
	t.Helper()
	w := httptest.NewRecorder()
	f, ok := catalogFilterFrom(w, httptest.NewRequest("GET", "/catalog?"+query, nil))
	if !ok {
		t.Fatalf("filter %q rejected: %s", query, w.Body)
	}
	return f
}

func TestCatalogFilterKeep(t *testing.T) {
	cases := []struct {
		query string
		want  []string
	}{
		{"", []string{"linux", "docker", "sql"}},
		{"spec=devops", []string{"linux", "docker"}},
		{"difficulty=beginner", []string{"linux", "sql"}},
		{"status=in_progress", []string{"docker"}},
		{"q=контейнеры", []string{"docker"}},
		{"q=BASH", []string{"linux"}},
		{"q=linux", []string{"linux"}},
		{"spec=devops&difficulty=beginner", []string{"linux"}},
		{"q=nothing+here", nil},
	}
	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			got := []string{}
			for _, course := range filterFrom(t, c.query).keep(filterCourses()) {
				got = append(got, course.Module.Slug)
			}
			if len(got) != len(c.want) {
				t.Fatalf("kept %v, want %v", got, c.want)
			}
			for i := range c.want {
				if got[i] != c.want[i] {
					t.Fatalf("kept %v, want %v", got, c.want)
				}
			}
		})
	}
}

func TestCatalogFilterRejectsUnknownValues(t *testing.T) {
	for _, query := range []string{"difficulty=impossible", "status=paused"} {
		w := httptest.NewRecorder()
		if _, ok := catalogFilterFrom(w, httptest.NewRequest("GET", "/catalog?"+query, nil)); ok {
			t.Fatalf("%q must be rejected", query)
		}
		if w.Code != 422 {
			t.Fatalf("%q status = %d, want 422", query, w.Code)
		}
	}
}

func TestCatalogFilterRejectsOverlongText(t *testing.T) {
	for _, query := range []string{
		"spec=" + strings.Repeat("a", catalogSpecMax+1),
		"q=" + strings.Repeat("я", catalogQueryMax+1),
	} {
		w := httptest.NewRecorder()
		if _, ok := catalogFilterFrom(w, httptest.NewRequest("GET", "/catalog?"+query, nil)); ok {
			t.Fatal("overlong value must be rejected")
		}
		if w.Code != 422 {
			t.Fatalf("status = %d, want 422", w.Code)
		}
	}
	// A value exactly at the limit still passes, even in multi-byte text.
	filterFrom(t, "q="+strings.Repeat("я", catalogQueryMax))
}
