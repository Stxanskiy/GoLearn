package handler

import "testing"

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
		if got := categorize(c.track, c.title, c.slug); got != c.want {
			t.Errorf("categorize(%q,%q,%q)=%q want %q", c.track, c.title, c.slug, got, c.want)
		}
	}
}

func TestHumanDuration(t *testing.T) {
	cases := []struct {
		min  int
		want string
	}{
		{0, ""}, {45, "~45 мин"}, {60, "~1 ч"}, {90, "~1 ч 30 мин"}, {120, "~2 ч"},
	}
	for _, c := range cases {
		if got := humanDuration(c.min); got != c.want {
			t.Errorf("humanDuration(%d)=%q want %q", c.min, got, c.want)
		}
	}
}

func TestDeriveLabel(t *testing.T) {
	cases := map[string]string{
		"beginner": "Старт", "intermediate": "Практика", "advanced": "Вызов", "expert": "Вызов", "": "Старт",
	}
	for in, want := range cases {
		if got := deriveLabel(in); got != want {
			t.Errorf("deriveLabel(%q)=%q want %q", in, got, want)
		}
	}
}

func TestLessonKindLabel(t *testing.T) {
	cases := map[string]string{"theory": "lesson", "quiz": "quiz", "lab": "lab", "sim": "sim", "sql": "sql", "": "lesson"}
	for kind, wantKey := range cases {
		if _, key := lessonKindLabel(kind); key != wantKey {
			t.Errorf("lessonKindLabel(%q) key=%q want %q", kind, key, wantKey)
		}
	}
}

func TestSpecForTrack(t *testing.T) {
	cases := map[string]string{
		"devops": "devops", "database": "database", "gym": "gym",
		// The Go track was removed from the platform, so anything unrecognised
		// now falls back to DevOps rather than to a specialization that is gone.
		"security": "security", "security-offense": "security", "golang": "devops", "weird": "devops",
	}
	for in, want := range cases {
		if got := specForTrack(in); got != want {
			t.Errorf("specForTrack(%q)=%q want %q", in, got, want)
		}
	}
}
