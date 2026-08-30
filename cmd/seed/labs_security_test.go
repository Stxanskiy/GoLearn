package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestSecurityLabsConverted verifies applySecurityLabs turns the security
// courses' legacy Go-coding tasks into shell labs: each mapped lesson ends up
// with shell tasks (no StarterCode), a setup on the first step, and check
// scripts that are valid bash. Reference solutions are verified separately by
// simulation.
func TestSecurityLabsConverted(t *testing.T) {
	mods := []M{mod_security_offense(), mod_security_defense()}
	applySecurityLabs(mods)

	covered := 0
	for mi := range mods {
		for li := range mods[mi].Lessons {
			l := &mods[mi].Lessons[li]
			lab, ok := securityLabs[l.Slug]
			if !ok {
				continue
			}
			covered++
			if len(l.Tasks) != len(lab.Steps) {
				t.Errorf("%s: %d tasks, want %d", l.Slug, len(l.Tasks), len(lab.Steps))
			}
			if len(l.Tasks) == 0 {
				continue
			}
			if l.Tasks[0].SetupScript == "" {
				t.Errorf("%s: first task has no SetupScript", l.Slug)
			}
			for i := range l.Tasks {
				task := &l.Tasks[i]
				if task.Kind != "shell" {
					t.Errorf("%s task %d: kind=%q, want shell", l.Slug, i+1, task.Kind)
				}
				if task.StarterCode != "" || len(task.TestCases) > 0 {
					t.Errorf("%s task %d: still has legacy Go code", l.Slug, i+1)
				}
				if strings.TrimSpace(task.CheckScript) == "" {
					t.Errorf("%s task %d: no CheckScript", l.Slug, i+1)
				}
				for _, sc := range []string{task.SetupScript, task.CheckScript} {
					if strings.TrimSpace(sc) == "" {
						continue
					}
					cmd := exec.Command("bash", "-n")
					cmd.Stdin = strings.NewReader(sc)
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Errorf("%s task %d: invalid bash: %v\n%s", l.Slug, i+1, err, out)
					}
				}
			}
		}
	}
	if covered != len(securityLabs) {
		t.Errorf("covered %d lessons, securityLabs has %d", covered, len(securityLabs))
	}
	t.Logf("converted %d security lessons to shell labs", covered)
}
