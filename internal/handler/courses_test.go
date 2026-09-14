package handler

import "testing"

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
