package api

import (
	"net/http"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
)

// In the storefront fixture student 1 has finished every published lesson of
// "linux" and is part-way through "docker", which makes them a ready-made pair:
// one prerequisite that is met and one that is not.
func prereqOn(t *testing.T, target string, ps ...model.Prerequisite) (*fakeUsers, *fakeContent) {
	t.Helper()
	users, c := storefront()
	// "helm" carries no lessons in the shared fixture, and a lesson request has
	// to reach the gate rather than stop at a 404 before it.
	c.lessons = append(c.lessons, model.Lesson{
		ID: 130, ModuleID: 13, Slug: "intro", Title: "Intro", Kind: "theory", OrderNum: 1, Published: true,
	})
	for i := range c.modules {
		if c.modules[i].Slug == target {
			c.modules[i].Prerequisites = ps
			return users, c
		}
	}
	t.Fatalf("no module %q in the fixture", target)
	return nil, nil
}

func TestPrerequisiteGate(t *testing.T) {
	req := func(slug string) model.Prerequisite {
		return model.Prerequisite{Slug: slug, Kind: model.PrereqRequired}
	}
	rec := func(slug string) model.Prerequisite {
		return model.Prerequisite{Slug: slug, Kind: model.PrereqRecommended}
	}

	cases := []struct {
		name       string
		prereqs    []model.Prerequisite
		wantLocked bool
	}{
		{"none", nil, false},
		{"required and finished", []model.Prerequisite{req("linux")}, false},
		{"required and unfinished", []model.Prerequisite{req("docker")}, true},
		{"recommended never locks", []model.Prerequisite{rec("docker")}, false},
		{"one unmet among met", []model.Prerequisite{req("linux"), req("docker")}, true},
		// A course naming itself, or a course that no longer exists, would lock
		// itself forever with nothing in the interface able to clear it.
		{"self reference ignored", []model.Prerequisite{req("helm")}, false},
		{"unknown course ignored", []model.Prerequisite{req("does-not-exist")}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			users, c := prereqOn(t, "helm", tc.prereqs...)
			h := newTestAPIWith(t, users, c)

			w := do(h, http.MethodGet, "/courses/helm", "", withCookie(studentToken))
			if w.Code != http.StatusOK {
				t.Fatalf("course page must stay open so the block can be explained: %d", w.Code)
			}
			got := decode[apigen.CourseDetail](t, w)
			if got.Locked != tc.wantLocked {
				t.Errorf("locked = %v, want %v", got.Locked, tc.wantLocked)
			}

			// The block is only worth anything if the content actually refuses.
			wl := do(h, http.MethodGet, "/courses/helm/lessons/intro", "", withCookie(studentToken))
			switch {
			case tc.wantLocked && wl.Code != http.StatusForbidden:
				t.Errorf("locked course served a lesson: %d", wl.Code)
			case !tc.wantLocked && wl.Code == http.StatusForbidden:
				t.Errorf("unlocked course refused a lesson: %d", wl.Code)
			}
		})
	}
}

// An author must be able to open the course they are writing without first
// sitting through its prerequisites.
func TestPrerequisiteGateSkipsAdmin(t *testing.T) {
	users, c := prereqOn(t, "helm", model.Prerequisite{Slug: "docker", Kind: model.PrereqRequired})
	h := newTestAPIWith(t, users, c)
	if w := do(h, http.MethodGet, "/courses/helm/lessons/intro", "", withCookie(adminToken)); w.Code == http.StatusForbidden {
		t.Errorf("admin was blocked by a prerequisite: %d", w.Code)
	}
}

// The column held a bare array of slugs before prerequisites had kinds. Those
// rows must keep working, and must block - that was their only meaning.
func TestPrerequisiteLegacyShape(t *testing.T) {
	var ps model.Prereqs
	if err := ps.UnmarshalJSON([]byte(`["docker","linux"]`)); err != nil {
		t.Fatalf("legacy shape: %v", err)
	}
	if len(ps) != 2 {
		t.Fatalf("got %d prerequisites, want 2", len(ps))
	}
	for _, p := range ps {
		if !p.Required() {
			t.Errorf("%s read as %q, want required", p.Slug, p.Kind)
		}
	}
	// null, an object, junk: unreadable prerequisites must not fail the course.
	for _, raw := range []string{`null`, `{"slug":"x"}`, `7`} {
		var out model.Prereqs
		if err := out.UnmarshalJSON([]byte(raw)); err != nil || len(out) != 0 {
			t.Errorf("%s → %v, %v; want empty and no error", raw, out, err)
		}
	}
}
