package repository

import (
	"context"
	"slices"
	"testing"

	"github.com/backendraz/golearn/internal/model"
)

// Every column of the table must be either carried into a draft or explicitly
// listed as not carried, with a reason.
//
// This is the test that was missing. prerequisites, vm_cpus and vm_mem_mib were
// each added to a table and to the studio, and nobody remembered the six column
// lists inside the draft SQL — so an author could set a prerequisite, send the
// draft for review, watch it approved, and find the requirement gone. Nothing
// failed; the work simply disappeared.
//
// Adding a column now breaks this test until someone says which it is.
func TestDraftColumnListsCoverTheTables(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()

	cases := []struct {
		table      string
		carried    []string
		notCarried map[string]string
	}{
		{"modules", draftModuleCols, moduleNotCarried},
		{"lessons", draftLessonCols, lessonNotCarried},
	}

	for _, tc := range cases {
		t.Run(tc.table, func(t *testing.T) {
			rows, err := pool.Query(ctx, `
				SELECT column_name FROM information_schema.columns
				 WHERE table_schema = 'public' AND table_name = $1
				 ORDER BY ordinal_position`, tc.table)
			if err != nil {
				t.Fatalf("read columns: %v", err)
			}
			defer rows.Close()

			var all []string
			for rows.Next() {
				var c string
				if err := rows.Scan(&c); err != nil {
					t.Fatal(err)
				}
				all = append(all, c)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if len(all) == 0 {
				t.Fatalf("table %s has no columns — wrong database?", tc.table)
			}

			for _, c := range all {
				_, excused := tc.notCarried[c]
				if !slices.Contains(tc.carried, c) && !excused {
					t.Errorf("%s.%s is in neither list: a draft would silently lose it.\n"+
						"    Add it to draft%sCols, or to %sNotCarried with the reason it stays behind.",
						tc.table, c, title(tc.table), tc.table[:len(tc.table)-1])
				}
			}
			// The other direction: a list naming a column that no longer exists
			// builds SQL that cannot run.
			for _, c := range tc.carried {
				if !slices.Contains(all, c) {
					t.Errorf("%s: carried column %q does not exist in the table", tc.table, c)
				}
			}
			for c := range tc.notCarried {
				if !slices.Contains(all, c) {
					t.Errorf("%s: %q is listed as not carried but does not exist", tc.table, c)
				}
			}
		})
	}
}

func title(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}

// What the author sets in a draft has to reach the live course when the draft
// is approved. The column lists above say which fields that is; this proves
// the SQL built from them actually moves them.
func TestDraftMergeCarriesTheAuthorsEdits(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	drafts := NewDraftRepo(pool)
	modules := NewModuleRepo(pool)
	lessons := NewLessonRepo(pool)

	live, err := modules.Create(ctx, model.Module{
		Slug: "draft-merge-test", Title: "Before", Track: "devops",
		Difficulty: "beginner", OrderNum: 9100, Published: true,
	})
	if err != nil {
		t.Fatalf("create course: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM modules WHERE id = $1 OR draft_of = $1`, live) })

	if _, err := lessons.Create(ctx, model.Lesson{
		ModuleID: live, Slug: "l1", Title: "Before", Kind: "lab", Format: "md",
		Difficulty: "beginner", OrderNum: 1, Published: true,
	}); err != nil {
		t.Fatalf("create lesson: %v", err)
	}

	draftID, err := drafts.Create(ctx, live)
	if err != nil {
		t.Fatalf("open draft: %v", err)
	}

	// Edit the draft the way the studio does: a prerequisite on the course and
	// a sandbox size on the lesson — the three fields that were being lost.
	if _, err := pool.Exec(ctx, `
		UPDATE modules SET title = 'After',
		  prerequisites = '[{"slug":"linux-start","kind":"required"}]'::jsonb
		 WHERE id = $1`, draftID); err != nil {
		t.Fatalf("edit draft course: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`UPDATE lessons SET title = 'After', vm_cpus = 2, vm_mem_mib = 1536 WHERE module_id = $1`,
		draftID); err != nil {
		t.Fatalf("edit draft lesson: %v", err)
	}

	// Approving a review is what merges a draft; call the merge directly so the
	// test stays about the columns rather than the review workflow.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := mergeDraft(ctx, tx, live); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("merge draft: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	got, err := modules.GetBySlug(ctx, "draft-merge-test")
	if err != nil {
		t.Fatalf("read course: %v", err)
	}
	if got.Title != "After" {
		t.Errorf("title = %q, want After", got.Title)
	}
	if len(got.Prerequisites) != 1 || got.Prerequisites[0].Slug != "linux-start" {
		t.Errorf("prerequisites did not survive the merge: %+v", got.Prerequisites)
	}

	ls, err := lessons.GetByModuleAll(ctx, live)
	if err != nil || len(ls) != 1 {
		t.Fatalf("read lessons: %v (%d)", err, len(ls))
	}
	if ls[0].Title != "After" {
		t.Errorf("lesson title = %q, want After", ls[0].Title)
	}
	if ls[0].VMCPUs != 2 || ls[0].VMMemMiB != 1536 {
		t.Errorf("sandbox size did not survive the merge: cpus=%d mem=%d", ls[0].VMCPUs, ls[0].VMMemMiB)
	}
}
