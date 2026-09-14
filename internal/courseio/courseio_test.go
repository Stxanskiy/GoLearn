package courseio

import (
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/model"
)

// TestRoundTrip checks that a course survives FromTree -> Marshal -> Parse -> ToTree
// unchanged for the fields that matter (including the 1-based correct mapping,
// order numbers and lab setup/check scripts).
func TestRoundTrip(t *testing.T) {
	orig := model.CourseTree{
		Module: model.Module{Slug: "k8s", Title: "Kube", Category: "Kubernetes", Difficulty: "advanced", Tags: []string{"k8s", "ckad"}},
		Lessons: []model.LessonBundle{
			{
				Lesson: model.Lesson{Slug: "l1", Title: "Intro", Kind: "theory", Format: "html", Content: "<p>hi</p>", OrderNum: 1},
				Questions: []model.QuizQuestion{
					{Question: "2+2?", Options: []string{"3", "4", "5"}, CorrectIndex: 1, Explanation: "four", OrderNum: 1},
				},
			},
			{
				Lesson: model.Lesson{Slug: "l2", Title: "Lab", Kind: "lab", VMImage: "golearn/sandbox-k8s", OrderNum: 2},
				Tasks: []model.Task{
					{Title: "apply", Kind: "shell", SandboxImage: "golearn/sandbox-k8s",
						SetupScript: "kubectl apply -f x.yaml", CheckScript: "kubectl get pod x", OrderNum: 1},
					{Title: "delete", Kind: "shell", CheckScript: "! kubectl get pod x", OrderNum: 2},
				},
			},
		},
	}

	data, err := Marshal(FromTree(orig))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := parsed.ToTree()

	if got.Module.Slug != "k8s" || got.Module.Title != "Kube" || got.Module.Category != "Kubernetes" {
		t.Errorf("module fields lost: %+v", got.Module)
	}
	if len(got.Module.Tags) != 2 || got.Module.Tags[0] != "k8s" {
		t.Errorf("tags lost: %v", got.Module.Tags)
	}
	if len(got.Lessons) != 2 {
		t.Fatalf("want 2 lessons, got %d", len(got.Lessons))
	}
	q := got.Lessons[0].Questions
	if len(q) != 1 || q[0].CorrectIndex != 1 || q[0].Options[1] != "4" {
		t.Errorf("quiz correct index round-trip broke: %+v", q)
	}
	if got.Lessons[0].Lesson.OrderNum != 1 || got.Lessons[1].Lesson.OrderNum != 2 {
		t.Errorf("lesson order lost")
	}
	tks := got.Lessons[1].Tasks
	if len(tks) != 2 {
		t.Fatalf("want 2 tasks, got %d", len(tks))
	}
	if tks[0].SetupScript != "kubectl apply -f x.yaml" || tks[0].CheckScript != "kubectl get pod x" {
		t.Errorf("lab scripts lost: %+v", tks[0])
	}
	if tks[0].SandboxImage != "golearn/sandbox-k8s" || tks[0].OrderNum != 1 || tks[1].OrderNum != 2 {
		t.Errorf("task image/order lost: %+v", tks)
	}
}

// TestValidate: duplicate/empty slugs are errors; odd kind/correct are warnings.
func TestValidate(t *testing.T) {
	c := Course{Slug: "s", Title: "t", Lessons: []Lesson{
		{Slug: "a", Title: "A", Kind: "sql"},
		{Slug: "a", Title: "dup"},
		{Slug: "", Title: "no slug"},
		{Slug: "Bad", Title: "B", Kind: "weird"},
		{Slug: "c", Title: "C", Quiz: []Question{{Question: "q", Options: []string{"x", "y"}, Correct: 9}},
			Tasks: []Task{{Title: "t", SandboxImage: "evil/image"}, {Title: "u", CheckScript: "true"}}},
	}}
	var got []string
	for _, is := range Validate(c) {
		got = append(got, is.Level+":"+is.Code+"@"+is.Path)
	}
	want := []string{
		"error:lesson_slug_duplicate@/lessons/1/slug",
		"error:lesson_slug_empty@/lessons/2/slug",
		"error:lesson_slug_invalid@/lessons/3/slug",
		"error:unknown_kind@/lessons/3/kind",
		"error:correct_out_of_range@/lessons/4/quiz/0/correct",
		"error:unknown_sandbox_image@/lessons/4/tasks/0/sandbox_image",
		"warn:check_without_image@/lessons/4/tasks/1/check_script",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("issues:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if HasErrors(Validate(Course{Slug: "s", Title: "t", Lessons: []Lesson{{Slug: "ok", Title: "ok", Kind: "lab"}}})) {
		t.Errorf("clean course should have no errors")
	}
}

// TestCorrectClamp: a missing/zero correct must not underflow to -1.
func TestCorrectClamp(t *testing.T) {
	c := Course{Slug: "s", Title: "t", Lessons: []Lesson{
		{Slug: "l", Title: "l", Quiz: []Question{{Question: "q", Options: []string{"a", "b"}, Correct: 0}}},
	}}
	got := c.ToTree()
	if got.Lessons[0].Questions[0].CorrectIndex != 0 {
		t.Errorf("want clamped 0, got %d", got.Lessons[0].Questions[0].CorrectIndex)
	}
}
