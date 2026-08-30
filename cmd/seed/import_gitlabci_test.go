package main

import "testing"

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
