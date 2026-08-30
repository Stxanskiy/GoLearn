package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestAnsibleCourse verifies the authored Ansible course imports offline (no DB):
// 14 lessons, a quiz, and 6 labs whose fixtures attach with valid-bash scripts.
// The check LOGIC is verified separately by running the reference playbooks
// against local ansible.
func TestAnsibleCourse(t *testing.T) {
	m, err := buildModule(importSpec{Dir: "module_ansible", Slug: "ansible", Track: "devops", Difficulty: "intermediate", Category: "Ansible"})
	if err != nil {
		t.Fatalf("buildModule: %v", err)
	}
	if got := len(m.Lessons); got != 14 {
		t.Fatalf("ansible lessons = %d, want 14", got)
	}
	if m.Title == "" {
		t.Fatal("course title empty")
	}
	labs := map[string]bool{}
	for _, s := range []string{"lab1", "lab2", "lab3", "lab4", "lab5", "lab6"} {
		labs["ch-ansible-"+s] = false
	}
	var quizzes int
	for i := range m.Lessons {
		l := &m.Lessons[i]
		if len(l.Quiz) > 0 {
			quizzes++
		}
		if containsImg(l.Content) {
			t.Errorf("lesson %q contains <img>", l.Slug)
		}
		if _, ok := labs[l.Slug]; !ok {
			continue
		}
		labs[l.Slug] = true
		if len(l.Tasks) == 0 || l.Tasks[0].SetupScript == "" {
			t.Errorf("%s: no setup/tasks", l.Slug)
		}
		checks := 0
		for _, task := range l.Tasks {
			for _, script := range []string{task.SetupScript, task.CheckScript} {
				if strings.TrimSpace(script) == "" {
					continue
				}
				cmd := exec.Command("bash", "-n")
				cmd.Stdin = strings.NewReader(script)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Errorf("%s: invalid bash: %v\n%s", l.Slug, err, out)
				}
			}
			if task.CheckScript != "" {
				checks++
			}
		}
		if checks == 0 {
			t.Errorf("%s: no auto-checks", l.Slug)
		}
	}
	if quizzes == 0 {
		t.Error("no quiz imported")
	}
	for slug, seen := range labs {
		if !seen {
			t.Errorf("lab %s not found", slug)
		}
	}
}
