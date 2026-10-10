package repository

import (
	"context"
	"testing"

	"github.com/backendraz/golearn/internal/model"
)

// Prerequisites survive a save and come back in the shape the gate reads.
//
// Every test of the gate itself runs against a fake store, so the one thing
// none of them can catch is the column: an Update that forgets the field, or a
// JSON shape the reader cannot parse, looks exactly like "the feature does not
// work" from the studio — the author sets a prerequisite, saves, and the course
// opens anyway.
func TestPrerequisitesRoundTrip(t *testing.T) {
	pool := billingPool(t)
	repo := NewModuleRepo(pool)
	ctx := context.Background()

	gate := model.Module{
		Slug: "prereq-gate-test", Title: "Gate", Track: "devops", Difficulty: "beginner",
		OrderNum: 9000, Published: false,
	}
	id, err := repo.Create(ctx, gate)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM modules WHERE id = $1`, id) })

	// A fresh course has none, and that must read as an empty list rather than
	// as something the reader chokes on.
	got, err := repo.GetBySlug(ctx, gate.Slug)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Prerequisites) != 0 {
		t.Fatalf("new course has prerequisites: %+v", got.Prerequisites)
	}

	got.Prerequisites = model.Prereqs{
		{Slug: "linux-start", Kind: model.PrereqRequired},
		{Slug: "git-basics", Kind: model.PrereqRecommended},
	}
	if err := repo.Update(ctx, *got); err != nil {
		t.Fatalf("update: %v", err)
	}

	back, err := repo.GetBySlug(ctx, gate.Slug)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if len(back.Prerequisites) != 2 {
		t.Fatalf("prerequisites = %+v, want 2", back.Prerequisites)
	}
	if back.Prerequisites[0].Slug != "linux-start" || !back.Prerequisites[0].Required() {
		t.Errorf("first = %+v, want linux-start required", back.Prerequisites[0])
	}
	// The kind is the whole difference between blocking and advising, so a
	// recommended one must not come back as required.
	if back.Prerequisites[1].Slug != "git-basics" || back.Prerequisites[1].Required() {
		t.Errorf("second = %+v, want git-basics recommended", back.Prerequisites[1])
	}

	// Clearing them has to work too: an author removing the last prerequisite
	// must not leave the course gated for ever.
	back.Prerequisites = model.Prereqs{}
	if err := repo.Update(ctx, *back); err != nil {
		t.Fatalf("clear: %v", err)
	}
	cleared, err := repo.GetBySlug(ctx, gate.Slug)
	if err != nil {
		t.Fatalf("get after clear: %v", err)
	}
	if len(cleared.Prerequisites) != 0 {
		t.Errorf("prerequisites survived the clear: %+v", cleared.Prerequisites)
	}
}
