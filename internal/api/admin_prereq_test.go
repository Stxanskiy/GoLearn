package api

import (
	"net/http"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
)

// Setting a prerequisite through the studio has to survive the round trip.
// The repository test covers the column; this covers everything between the
// request body and it — validation, the "omitted means unchanged" rule, and
// the read back.
func TestAdminCoursePrerequisitesSurviveUpdate(t *testing.T) {
	h, _ := authorsFixture(t)

	read := func() apigen.AdminCourse {
		t.Helper()
		w := do(h, http.MethodGet, "/admin/courses/12", "", withCookie(coToken))
		if w.Code != http.StatusOK {
			t.Fatalf("read: %d %s", w.Code, w.Body)
		}
		return decode[apigen.AdminCourseDetail](t, w).Course
	}

	before := read()
	base := `{"slug":"` + before.Slug + `","title":"` + before.Title +
		`","track":"` + before.Track + `","difficulty":"beginner"`

	// Adding one.
	body := base + `,"prerequisites":[{"slug":"linux","kind":"required"}]}`
	if w := do(h, http.MethodPut, "/admin/courses/12", body, withCookie(coToken)); w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body)
	}
	got := read()
	if len(got.Prerequisites) != 1 {
		t.Fatalf("after adding one: %+v", got.Prerequisites)
	}
	if got.Prerequisites[0].Slug != "linux" || got.Prerequisites[0].Kind != apigen.Required {
		t.Fatalf("stored = %+v", got.Prerequisites[0])
	}

	// Omitting the field must leave it alone — the studio sends a full body, but
	// anything else talking to this endpoint must not wipe it by silence.
	if w := do(h, http.MethodPut, "/admin/courses/12", base+`}`, withCookie(coToken)); w.Code != http.StatusOK {
		t.Fatalf("update without the field: %d %s", w.Code, w.Body)
	}
	if got = read(); len(got.Prerequisites) != 1 {
		t.Fatalf("omitting the field wiped it: %+v", got.Prerequisites)
	}

	// Clearing it explicitly must work.
	if w := do(h, http.MethodPut, "/admin/courses/12", base+`,"prerequisites":[]}`, withCookie(coToken)); w.Code != http.StatusOK {
		t.Fatalf("clear: %d %s", w.Code, w.Body)
	}
	if got = read(); len(got.Prerequisites) != 0 {
		t.Fatalf("clearing left %+v", got.Prerequisites)
	}
}
