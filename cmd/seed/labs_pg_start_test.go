package main

import (
	"os/exec"
	"strings"
	"testing"
)

// The PostgreSQL course is written against its sandbox; these hold it to the
// rules that sandbox and the labcheck harness rely on.
func TestPgStartLabs(t *testing.T) {
	var mod *M
	mods := importedModules()
	for i := range mods {
		if mods[i].Slug == "pg-start" {
			mod = &mods[i]
		}
	}
	if mod == nil {
		t.Fatal("pg-start is not imported")
	}
	// pg-start is finished and published; the rest of the path is not. pg-sql,
	// pg-ops and gym-sql are still chapter headings - around 210 characters of
	// body each - and a deploy must not put them in front of students, so the
	// guard stays on them and only lifts for the course that is done.
	for _, m := range mods {
		switch m.Slug {
		case "pg-sql", "pg-ops", "gym-sql":
			if !m.Draft {
				t.Errorf("%s must stay a draft until an admin publishes it", m.Slug)
			}
		case "pg-start":
			if m.Draft {
				t.Error("pg-start is published; it must not be seeded as a draft")
			}
		}
	}
	lessons := map[string]L{}
	keys := map[string]string{}
	for _, l := range mod.Lessons {
		lessons[l.Slug] = l
		for _, task := range l.Tasks {
			if task.SourceKey == "" {
				t.Errorf("%s: task %q has no taskid", l.Slug, task.Title)
			} else if prev, dup := keys[task.SourceKey]; dup {
				t.Errorf("taskid %s used in %s and %s", task.SourceKey, prev, l.Slug)
			}
			keys[task.SourceKey] = l.Slug
		}
	}
	for slug, spec := range pgStartLabs {
		l, ok := lessons[slug]
		if !ok {
			t.Errorf("fixture %s has no lesson in course_pg_start", slug)
			continue
		}
		checked := 0
		for _, task := range l.Tasks {
			if task.SelfCheck {
				t.Errorf("%s: self-check task %q — every task of this course is checked", slug, task.Title)
				continue
			}
			checked++
			if task.SandboxImage != sandboxImagePG {
				t.Errorf("%s: task %q runs on %s, not %s", slug, task.Title, task.SandboxImage, sandboxImagePG)
			}
			// labcheck flattens a check to one line before running it.
			if strings.Contains(task.CheckScript, "\n") {
				t.Errorf("%s: check of %q spans lines", slug, task.Title)
			}
		}
		for n := range spec.Checks {
			if n < 1 || n > checked {
				t.Errorf("%s: check #%d has no task (the lesson has %d)", slug, n, checked)
			}
		}
		for _, script := range append([]string{spec.Setup}, checksOf(spec)...) {
			cmd := exec.Command("bash", "-n")
			cmd.Stdin = strings.NewReader(script)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("%s: script does not parse: %v\n%s", slug, err, out)
			}
		}
	}
	for _, l := range mod.Lessons {
		if l.Kind == "lab" {
			if _, ok := pgStartLabs[l.Slug]; !ok {
				t.Errorf("lab %s has no fixture", l.Slug)
			}
		}
	}
}

func checksOf(s labSpec) []string {
	var out []string
	for _, c := range s.Checks {
		out = append(out, c)
	}
	return out
}
