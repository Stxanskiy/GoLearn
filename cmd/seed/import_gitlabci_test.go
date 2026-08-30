package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestGitlabCIImports verifies the imported GitLab CI/CD course parses from the
// embedded devops404 export: 20 chapters, with quizzes and lab (shell) lessons.
// It needs no database — buildModule reads only the embedded content FS.
func TestGitlabCIImports(t *testing.T) {
	m, err := buildModule(importSpec{
		Dir: "crs_gitlab_ci", Slug: "gitlab-ci", Track: "devops",
		Difficulty: "intermediate", Category: "CI/CD",
	})
	if err != nil {
		t.Fatalf("buildModule: %v", err)
	}
	if got := len(m.Lessons); got != 20 {
		t.Fatalf("lessons = %d, want 20", got)
	}
	if m.Title == "" {
		t.Fatalf("course title is empty")
	}
	var quizzes, labs int
	for _, l := range m.Lessons {
		if len(l.Quiz) > 0 {
			quizzes++
		}
		if len(l.Tasks) > 0 {
			labs++
		}
		// cleanContent must have stripped the embedded cover images.
		if containsImg(l.Content) {
			t.Errorf("lesson %q still contains an <img> tag", l.Slug)
		}
	}
	if quizzes == 0 {
		t.Errorf("no quizzes imported")
	}
	if labs == 0 {
		t.Errorf("no lab (shell) lessons imported")
	}
	t.Logf("gitlab-ci: %d lessons, %d with quizzes, %d with lab tasks", len(m.Lessons), quizzes, labs)
}

func containsImg(s string) bool {
	return imgTag.MatchString(s)
}

// TestGitlabCILab1Fixtures checks that lab 1's fixtures attach and that every
// generated Setup/Check script is syntactically valid bash (bash -n). It does
// not run the checks (that needs a live sandbox) — only that they parse.
func TestGitlabCILab1Fixtures(t *testing.T) {
	m, err := buildModule(importSpec{Dir: "crs_gitlab_ci", Slug: "gitlab-ci", Track: "devops", Difficulty: "intermediate", Category: "CI/CD"})
	if err != nil {
		t.Fatalf("buildModule: %v", err)
	}
	var lab *L
	for i := range m.Lessons {
		if m.Lessons[i].Slug == "ch-gitlab-ci-lab1" {
			lab = &m.Lessons[i]
			break
		}
	}
	if lab == nil {
		t.Fatal("lesson ch-gitlab-ci-lab1 not found")
	}
	if len(lab.Tasks) == 0 {
		t.Fatal("lab1 has no tasks")
	}
	if lab.Tasks[0].SetupScript == "" {
		t.Error("lab1 first task has no SetupScript")
	}
	checks := 0
	for i, task := range lab.Tasks {
		for _, script := range []string{task.SetupScript, task.CheckScript} {
			if strings.TrimSpace(script) == "" {
				continue
			}
			cmd := exec.Command("bash", "-n")
			cmd.Stdin = strings.NewReader(script)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("task %d script is not valid bash: %v\n%s\n---\n%s", i+1, err, out, script)
			}
		}
		if task.CheckScript != "" {
			checks++
		}
	}
	if checks == 0 {
		t.Error("lab1 has no auto-checks attached")
	}
	t.Logf("lab1: %d tasks, %d with auto-checks", len(lab.Tasks), checks)
}
