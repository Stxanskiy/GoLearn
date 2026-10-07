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
	if !mod.Draft {
		t.Error("pg-start must stay a draft until it is published by an admin")
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
