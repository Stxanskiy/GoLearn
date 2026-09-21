package catalog

import (
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/model"
)

func TestSortByProgress(t *testing.T) {
	now := time.Now()
	courses := []Course{
		{Module: model.Module{Slug: "done"}, Status: StatusCompleted},
		{Module: model.Module{Slug: "fresh"}, Status: StatusInProgress, LastActivity: now},
		{Module: model.Module{Slug: "new-a"}, Status: StatusNotStarted},
		{Module: model.Module{Slug: "stale"}, Status: StatusInProgress, LastActivity: now.Add(-48 * time.Hour)},
		{Module: model.Module{Slug: "new-b"}, Status: StatusNotStarted},
	}
	SortByProgress(courses)

	want := []string{"fresh", "stale", "new-a", "new-b", "done"}
	for i, slug := range want {
		if courses[i].Module.Slug != slug {
			got := make([]string, len(courses))
			for j, c := range courses {
				got[j] = c.Module.Slug
			}
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestBuildCourseLastActivity(t *testing.T) {
	newest := time.Now()
	lessons := []model.Lesson{{ID: 1}, {ID: 2}}
	up := NewUserProgress([]model.Progress{
		{LessonID: 1, Status: StatusCompleted, UpdatedAt: newest.Add(-time.Hour)},
		{LessonID: 2, Status: StatusInProgress, UpdatedAt: newest},
	}, nil)

	c := BuildCourse(model.Module{Slug: "x"}, lessons, up)
	if !c.LastActivity.Equal(newest) {
		t.Fatalf("last activity = %v, want %v", c.LastActivity, newest)
	}
	if c.Status != StatusInProgress {
		t.Fatalf("status = %q", c.Status)
	}
}
