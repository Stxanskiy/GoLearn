package main

import (
	"sort"
	"testing"
)

// The source keeps a chapter's questions and tasks in one ordered list and the
// author interleaves them deliberately - a task, then the question that checks
// it, then the next task. buildModule splits that list in two, and before Pos
// existed each half was numbered from one, so the lesson came out as every
// question followed by every work step.
//
// This holds the split to the source: rebuilding the sequence from the two
// halves by position must give back exactly what the author wrote.
func TestLessonStepsKeepTheAuthorsOrder(t *testing.T) {
	mods := importedModules()
	if len(mods) == 0 {
		t.Fatal("no modules imported")
	}

	checked, interleaved := 0, 0
	for _, m := range mods {
		for _, l := range m.Lessons {
			if len(l.Quiz) == 0 || len(l.Tasks) == 0 {
				continue // nothing to interleave
			}
			checked++

			// Every step must have a position, or the order is guesswork.
			seen := map[int]string{}
			for _, q := range l.Quiz {
				if q.Pos == 0 {
					t.Errorf("%s/%s: question %q has no position", m.Slug, l.Slug, q.Question[:min(30, len(q.Question))])
					continue
				}
				if prev, dup := seen[q.Pos]; dup {
					t.Errorf("%s/%s: position %d used by both %q and a question", m.Slug, l.Slug, q.Pos, prev)
				}
				seen[q.Pos] = "question"
			}
			for _, task := range l.Tasks {
				if task.Pos == 0 {
					t.Errorf("%s/%s: task %q has no position", m.Slug, l.Slug, task.Title)
					continue
				}
				if _, dup := seen[task.Pos]; dup {
					t.Errorf("%s/%s: position %d used twice (task %q)", m.Slug, l.Slug, task.Pos, task.Title)
				}
				seen[task.Pos] = "task"
			}

			// Gaps are legitimate: the source carries step types the importer does
			// not build - there is one "playbook" in the whole content - and those
			// leave their position empty. Order is what matters, not a dense run,
			// so read the positions that exist, in order.
			order := make([]int, 0, len(seen))
			for at := range seen {
				order = append(order, at)
			}
			sort.Ints(order)

			seq := ""
			for _, at := range order {
				if seen[at] == "question" {
					seq += "Q"
				} else {
					seq += "T"
				}
			}
			if sorted := sortedByKind(seq); seq != sorted {
				interleaved++
			}
		}
	}

	if checked == 0 {
		t.Fatal("no lesson carries both questions and tasks - the fixture cannot prove anything")
	}
	// If nothing is interleaved the split is free and this test proves nothing,
	// which is exactly the state the bug left the content in.
	if interleaved == 0 {
		t.Error("no lesson interleaves questions and tasks; the author's order is not being carried")
	}
	t.Logf("%d lessons carry both; %d of them interleave", checked, interleaved)
}

// sortedByKind is the sequence the old importer produced: every question, then
// every task.
func sortedByKind(seq string) string {
	out := ""
	for _, c := range seq {
		if c == 'Q' {
			out += "Q"
		}
	}
	for _, c := range seq {
		if c != 'Q' {
			out += string(c)
		}
	}
	return out
}
