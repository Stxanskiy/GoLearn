package catalog

import (
	"slices"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/model"
)

func TestCategorize(t *testing.T) {
	cases := []struct{ track, title, slug, want string }{
		{"golang", "Первые шаги в Go", "basics", "Golang"},
		{"devops", "Kubernetes — Основы", "k8s-intro", "Kubernetes"},
		{"devops", "Docker: команды", "docker-basics", "Docker"},
		{"devops", "Linux: Старт", "linux-start", "Linux"},
		{"devops", "Git: Основы", "git-basics", "Git"},
		{"database", "Экспресс курс по SQL", "sql-express", "Database"},
		{"security", "Пентест", "sec-off", "Security"},
		{"devops", "Helm", "helm", "Kubernetes"},
		{"devops", "CI/CD культура", "devops", "DevOps"},
	}
	for _, c := range cases {
		if got := Categorize(c.track, c.title, c.slug); got != c.want {
			t.Errorf("Categorize(%q,%q,%q)=%q want %q", c.track, c.title, c.slug, got, c.want)
		}
	}
	if got := Category(model.Module{Category: "Custom", Title: "Docker"}); got != "Custom" {
		t.Errorf("explicit category ignored: %q", got)
	}
}

func TestSpecForTrack(t *testing.T) {
	cases := map[string]string{
		"devops": "devops", "database": "database", "gym": "gym",
		"security": "security", "security-offense": "security", "backend": "devops", "": "devops", "frontend": "frontend",
	}
	for in, want := range cases {
		if got := SpecForTrack(in); got != want {
			t.Errorf("SpecForTrack(%q)=%q want %q", in, got, want)
		}
		if !slices.Contains(SpecTracks(want), in) {
			t.Errorf("SpecTracks(%q) does not contain %q", want, in)
		}
	}
}

func TestLabelCode(t *testing.T) {
	cases := []struct {
		m    model.Module
		want string
	}{
		{model.Module{Label: "Вызов", Difficulty: "beginner"}, LabelChallenge},
		{model.Module{Label: "practice"}, LabelPractice},
		{model.Module{Difficulty: "intermediate"}, LabelPractice},
		{model.Module{Difficulty: "expert"}, LabelChallenge},
		{model.Module{Label: "unknown"}, LabelStart},
	}
	for _, c := range cases {
		if got := LabelCode(c.m); got != c.want {
			t.Errorf("LabelCode(%+v)=%q want %q", c.m, got, c.want)
		}
	}
}

func TestLessonStatus(t *testing.T) {
	score := 3
	up := NewUserProgress([]model.Progress{
		{LessonID: 1, Status: StatusCompleted},
		{LessonID: 2, Status: StatusCompleted}, // quiz marked done without a score
		{LessonID: 3, Status: StatusInProgress, QuizScore: &score},
		{LessonID: 4, Status: StatusCompleted},  // lab marked done, tasks not passed
		{LessonID: 6, Status: StatusNotStarted}, // quiz after retry
	}, map[int]bool{5: true})

	cases := []struct {
		lesson model.Lesson
		want   string
	}{
		{model.Lesson{ID: 1, Kind: "theory"}, StatusCompleted},
		{model.Lesson{ID: 2, Kind: "quiz"}, StatusInProgress},
		{model.Lesson{ID: 3, Kind: "quiz"}, StatusCompleted},
		{model.Lesson{ID: 4, Kind: "lab"}, StatusInProgress},
		{model.Lesson{ID: 5, Kind: "lab"}, StatusCompleted},
		{model.Lesson{ID: 6, Kind: "quiz"}, StatusNotStarted},
		{model.Lesson{ID: 7, Kind: "theory"}, StatusNotStarted},
	}
	for _, c := range cases {
		if got := up.LessonStatus(c.lesson); got != c.want {
			t.Errorf("lesson %d (%s): %q want %q", c.lesson.ID, c.lesson.Kind, got, c.want)
		}
	}
}

func TestBuildCourse(t *testing.T) {
	m := model.Module{Slug: "docker", Title: "Docker", Track: "devops"}
	lessons := []model.Lesson{{ID: 1, Kind: "theory"}, {ID: 2, Kind: "theory"}, {ID: 3, Kind: "lab"}}

	empty := BuildCourse(m, lessons, NewUserProgress(nil, nil))
	if empty.Status != StatusNotStarted || empty.Pct != 0 || empty.EstMinutes != 30 || empty.NextLesson() != 0 {
		t.Errorf("empty progress: %+v", empty)
	}

	up := NewUserProgress([]model.Progress{{LessonID: 1, Status: StatusCompleted}, {LessonID: 2, Status: StatusInProgress}}, nil)
	c := BuildCourse(m, lessons, up)
	if c.Status != StatusInProgress || c.Completed != 1 || c.Pct != 33 || c.NextLesson() != 1 || c.Category != "Docker" || c.Spec != "devops" {
		t.Errorf("partial progress: %+v", c)
	}

	up = NewUserProgress([]model.Progress{{LessonID: 1, Status: StatusCompleted}, {LessonID: 2, Status: StatusCompleted}}, map[int]bool{3: true})
	if c := BuildCourse(m, lessons, up); c.Status != StatusCompleted || c.Pct != 100 || c.NextLesson() != -1 {
		t.Errorf("all done: %+v", c)
	}

	if c := BuildCourse(m, nil, up); c.Status != StatusNotStarted || c.Pct != 0 {
		t.Errorf("no lessons: %+v", c)
	}
}

func TestCoverSVGEscapes(t *testing.T) {
	spec := SpecCoverSVG(model.Specialization{Slug: "x", Icon: `<script>`})
	if strings.Contains(spec, "<script>") {
		t.Errorf("spec icon not escaped: %s", spec)
	}
	course := CourseCoverSVG(model.Module{Category: `A&B<`})
	if strings.Contains(course, "A&B<") || !strings.Contains(course, "A&amp;B&lt;") {
		t.Errorf("category not escaped: %s", course)
	}
}

func TestDecodeDataURI(t *testing.T) {
	mime, data, ok := DecodeDataURI("data:image/png;base64,aGk=")
	if !ok || mime != "image/png" || string(data) != "hi" {
		t.Errorf("got %q %q %v", mime, data, ok)
	}
	if _, _, ok := DecodeDataURI("https://x/y.png"); ok {
		t.Error("URL must not decode")
	}
}
