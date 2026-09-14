package api

import (
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
)

// lessonsFixture extends the storefront with lesson content, a quiz and a lab.
func lessonsFixture(t *testing.T) (http.Handler, *fakeContent) {
	t.Helper()
	users, c := storefront()
	schema := `{"title":"Shop","tables":[{"id":"orders","cols":[{"name":"id","type":"INTEGER","isKey":true,"desc":"PK"}]}],"ddl":"CREATE TABLE orders (id INTEGER);","fields":["id"]}`
	for i := range c.lessons {
		switch c.lessons[i].ID {
		case 100:
			c.lessons[i].Format, c.lessons[i].Content = "md", "# Intro\n\n<script>alert(1)</script>"
		case 110:
			c.lessons[i].Kind = "sql"
			c.lessons[i].Content = `<p>Select ids</p><div class="sql-erd" data-schema="` + html.EscapeString(schema) + `"></div>`
		}
	}
	c.lessons = append(c.lessons, model.Lesson{ID: 130, ModuleID: 13, Slug: "charts", Title: "Charts", Kind: "theory", OrderNum: 1, Published: true})
	c.questions[101] = []model.QuizQuestion{
		{ID: 1001, Question: "Pick <b>two</b>", Options: []string{"one", "two", "three"}, OptionExpl: []string{"no", "yes", "no"}, CorrectIndex: 1, Explanation: "Because"},
		{ID: 1002, Question: `<img src=x onerror="steal()">Yes?`, Options: []string{"yes", "no"}, CorrectIndex: 0},
	}
	c.tasks[102] = 3
	return newTestAPIWith(t, users, c), c
}

func TestGetLesson(t *testing.T) {
	h, _ := lessonsFixture(t)

	w := do(h, http.MethodGet, "/courses/linux/lessons/quiz", "", withCookie(studentToken))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	body := w.Body.String()
	for _, leak := range []string{"correct_index", "Because", `"yes","no","no"`, "onerror"} {
		if strings.Contains(body, leak) {
			t.Errorf("unanswered quiz leaks %q: %s", leak, body)
		}
	}
	l := decode[apigen.LessonDetail](t, w)
	if l.Kind != "quiz" || l.HasLab || l.Quiz == nil || len(l.Quiz.Questions) != 2 {
		t.Fatalf("lesson = %+v", l)
	}
	if l.Quiz.Questions[0].QuestionHTML != "Pick <b>two</b>" || l.Quiz.Questions[0].Answer != nil {
		t.Errorf("question = %+v", l.Quiz.Questions[0])
	}
	nav := l.Nav
	if nav.Position != 2 || nav.Total != 3 || nav.Prev == nil || nav.Prev.Slug != "intro" || nav.Next == nil || nav.Next.Kind != "lab" || nav.CourseProgressPct != 100 {
		t.Errorf("nav = %+v prev %+v next %+v", nav, nav.Prev, nav.Next)
	}
	if l.Progress.Status != "completed" || l.Progress.QuizScore == nil {
		t.Errorf("progress = %+v", l.Progress)
	}

	l = decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/linux/lessons/intro", "", withCookie(studentToken)))
	if !strings.Contains(l.ContentHTML, "<h1") || strings.Contains(l.ContentHTML, "<script") || l.Quiz != nil || l.SQLSchema != nil {
		t.Errorf("intro = %+v", l)
	}
	if l := decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/linux/lessons/lab", "", withCookie(studentToken))); !l.HasLab {
		t.Error("lab lesson must report has_lab")
	}

	l = decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/docker/lessons/images", "", withCookie(studentToken)))
	if s := l.SQLSchema; s == nil || s.Title != "Shop" || len(s.Tables) != 1 || !s.Tables[0].Columns[0].IsKey || s.ExpectedColumns[0] != "id" || s.Ddl == "" {
		t.Errorf("sql schema = %+v", l.SQLSchema)
	}

	for _, tt := range []struct {
		path, token string
		want        int
	}{
		{"/courses/linux/lessons/draft", studentToken, http.StatusNotFound},
		{"/courses/k8s-draft/lessons/pods", studentToken, http.StatusNotFound},
		{"/courses/linux/lessons/nope", studentToken, http.StatusNotFound},
		{"/courses/nope/lessons/intro", studentToken, http.StatusNotFound},
		{"/courses/k8s-draft/lessons/pods", adminToken, http.StatusOK},
	} {
		if w := do(h, http.MethodGet, tt.path, "", withCookie(tt.token)); w.Code != tt.want {
			t.Errorf("%s: %d want %d", tt.path, w.Code, tt.want)
		}
	}
	l = decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/linux/lessons/draft", "", withCookie(adminToken)))
	if l.Nav.Position != 0 || l.Nav.Prev != nil || l.Nav.Next != nil {
		t.Errorf("draft lesson nav = %+v", l.Nav)
	}
}

func TestVisitCompleteAndNotes(t *testing.T) {
	h, _ := lessonsFixture(t)
	post := func(path, body string) *httptest.ResponseRecorder {
		return do(h, http.MethodPost, path, body, withCookie(studentToken))
	}

	w := post("/lessons/130/visit", "")
	if w.Code != http.StatusOK || decode[apigen.LessonProgress](t, w).Status != "in_progress" {
		t.Fatalf("visit: %d %s", w.Code, w.Body)
	}
	w = post("/lessons/130/complete", "")
	if w.Code != http.StatusOK || decode[apigen.LessonProgress](t, w).Status != "completed" {
		t.Fatalf("complete: %d %s", w.Code, w.Body)
	}
	if p := decode[apigen.LessonProgress](t, post("/lessons/130/visit", "")); p.Status != "completed" {
		t.Errorf("visit must not reset a completed lesson: %+v", p)
	}

	for _, id := range []string{"101", "102"} {
		if w := post("/lessons/"+id+"/complete", ""); w.Code != http.StatusConflict || errorCode(t, w) != codeCompletionDerived {
			t.Errorf("complete %s: %d %s", id, w.Code, w.Body)
		}
	}
	for _, path := range []string{"/lessons/abc/visit", "/lessons/999/visit", "/lessons/103/visit", "/lessons/120/complete"} {
		if w := post(path, ""); w.Code != http.StatusNotFound {
			t.Errorf("%s: %d", path, w.Code)
		}
	}

	put := func(body string) *httptest.ResponseRecorder {
		return do(h, http.MethodPut, "/lessons/130/notes", body, withCookie(studentToken))
	}
	if w := put(`{"notes":"remember -la"}`); w.Code != http.StatusNoContent {
		t.Fatalf("notes: %d %s", w.Code, w.Body)
	}
	if w := put(`{"notes":"` + strings.Repeat("я", maxNotesLen+1) + `"}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("long notes: %d", w.Code)
	}
	l := decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/helm/lessons/charts", "", withCookie(studentToken)))
	if l.Progress.Notes != "remember -la" || l.Progress.Status != "completed" {
		t.Errorf("progress = %+v", l.Progress)
	}
}

func TestQuizFlow(t *testing.T) {
	h, c := lessonsFixture(t)
	c.progress[1] = nil // fresh attempt for the student
	post := func(path, body string) *httptest.ResponseRecorder {
		return do(h, http.MethodPost, path, body, withCookie(studentToken))
	}

	w := post("/lessons/101/quiz/answers", `{"question_id":1001,"selected_index":0}`)
	if w.Code != http.StatusOK {
		t.Fatalf("answer: %d %s", w.Code, w.Body)
	}
	a := decode[apigen.QuizAnswerResult](t, w)
	if a.IsCorrect || a.CorrectIndex != 1 || a.Locked || a.OptionExplanationsHTML == nil || (*a.OptionExplanationsHTML)[1] != "yes" || a.ExplanationHTML == nil {
		t.Errorf("first answer = %+v", a)
	}
	if a := decode[apigen.QuizAnswerResult](t, post("/lessons/101/quiz/answers", `{"question_id":1001,"selected_index":1}`)); !a.Locked || a.SelectedIndex != 0 || a.IsCorrect {
		t.Errorf("repeat answer must keep the first one: %+v", a)
	}

	l := decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/linux/lessons/quiz", "", withCookie(studentToken)))
	if l.Quiz.Questions[0].Answer == nil || l.Quiz.Questions[0].Answer.CorrectIndex != 1 || l.Quiz.Questions[1].Answer != nil {
		t.Errorf("answered state = %+v / %+v", l.Quiz.Questions[0].Answer, l.Quiz.Questions[1].Answer)
	}
	if strings.Contains(l.Quiz.Questions[1].QuestionHTML, "onerror") {
		t.Errorf("question html not sanitized: %q", l.Quiz.Questions[1].QuestionHTML)
	}

	if w := post("/lessons/101/quiz/answers", `{"question_id":9999,"selected_index":0}`); w.Code != http.StatusNotFound {
		t.Errorf("foreign question: %d", w.Code)
	}
	if w := post("/lessons/101/quiz/answers", `{"question_id":1002,"selected_index":2}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("out of range: %d", w.Code)
	}
	if w := post("/lessons/101/quiz/submit", `{"answers":[{"question_id":1002,"selected_index":5}]}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("submit out of range: %d", w.Code)
	}

	w = post("/lessons/101/quiz/submit", `{"answers":[{"question_id":1001,"selected_index":1},{"question_id":1002,"selected_index":0}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("submit: %d %s", w.Code, w.Body)
	}
	res := decode[apigen.QuizResult](t, w)
	if res.Score != 1 || res.Total != 2 || res.Percent != 50 || *res.Results[0].SelectedIndex != 0 || !res.Results[1].IsCorrect {
		t.Errorf("result = %+v", res)
	}
	if l := decode[apigen.LessonDetail](t, do(h, http.MethodGet, "/courses/linux/lessons/quiz", "", withCookie(studentToken))); l.Progress.Status != "completed" || *l.Progress.QuizScore != 1 {
		t.Errorf("progress after submit = %+v", l.Progress)
	}

	if w := do(h, http.MethodDelete, "/lessons/101/quiz/answers", "", withCookie(studentToken)); w.Code != http.StatusNoContent {
		t.Fatalf("reset: %d", w.Code)
	}
	res = decode[apigen.QuizResult](t, post("/lessons/101/quiz/submit", ""))
	if res.Score != 0 || res.Results[0].SelectedIndex != nil || res.Results[1].IsCorrect {
		t.Errorf("empty submit after reset = %+v", res)
	}

	if w := post("/lessons/100/quiz/submit", ""); w.Code != http.StatusNotFound {
		t.Errorf("submit without quiz: %d", w.Code)
	}
}
