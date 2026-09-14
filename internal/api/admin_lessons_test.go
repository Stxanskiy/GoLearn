package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
)

func TestAdminLessons(t *testing.T) {
	h, c := authorsFixture(t)
	req := func(token, method, path, body string) *httptest.ResponseRecorder {
		return do(h, method, path, body, withCookie(token))
	}
	lesson := `{"slug":"services","title":"Services","kind":"lab","format":"md","difficulty":"intermediate","content":"# S","vm_image":"kuber","vm_init":"k8s-start"}`

	if w := req(coToken, http.MethodPost, "/admin/courses/12/lessons", strings.Replace(lesson, `"slug"`, `"published":true,"slug"`, 1)); w.Code != http.StatusForbidden {
		t.Errorf("co-author creates published lesson: %d", w.Code)
	}
	w := req(coToken, http.MethodPost, "/admin/courses/12/lessons", lesson)
	if w.Code != http.StatusCreated {
		t.Fatalf("create lesson: %d %s", w.Code, w.Body)
	}
	created := decode[apigen.AdminLesson](t, w)
	if created.Published || created.OrderNum != 2 || created.CourseID != 12 || created.VMInit != "k8s-start" || created.Source != "admin" {
		t.Errorf("created = %+v", created)
	}
	if w := req(coToken, http.MethodPost, "/admin/courses/12/lessons", lesson); w.Code != http.StatusConflict || errorCode(t, w) != codeSlugTaken {
		t.Errorf("duplicate lesson slug: %d %s", w.Code, w.Body)
	}
	if w := req(coToken, http.MethodPost, "/admin/courses/12/lessons", `{"slug":"x","title":"X","kind":"video","format":"md","difficulty":"beginner"}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("unknown kind: %d", w.Code)
	}
	if w := req(otherToken, http.MethodPost, "/admin/courses/12/lessons", lesson); w.Code != http.StatusNotFound {
		t.Errorf("unrelated author: %d", w.Code)
	}
	path := "/admin/lessons/" + itoa(created.ID)

	question := `{"question":"Which?","options":["a","b","c"],"option_explanations":["no","yes","no"],"correct_index":1,"explanation":"b"}`
	if w := req(coToken, http.MethodPost, path+"/questions", `{"question":"Q","options":["a","b"],"option_explanations":["x"],"correct_index":2}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("invalid question: %d %s", w.Code, w.Body)
	} else if e := decode[apigen.Error](t, w); (*e.Error.Details.Fields)["correct_index"] != fieldInvalidValue || (*e.Error.Details.Fields)["option_explanations"] != fieldInvalidValue {
		t.Errorf("question fields = %v", *e.Error.Details.Fields)
	}
	w = req(coToken, http.MethodPost, path+"/questions", question)
	q := decode[apigen.AdminQuestion](t, w)
	if w.Code != http.StatusCreated || q.LessonID != created.ID || q.OrderNum != 1 || len(q.OptionExplanations) != 3 {
		t.Fatalf("create question: %d %+v", w.Code, q)
	}
	if w := req(coToken, http.MethodPut, "/admin/questions/"+itoa(q.ID), strings.Replace(question, "Which?", "Which one?", 1)); w.Code != http.StatusOK || decode[apigen.AdminQuestion](t, w).Question != "Which one?" {
		t.Errorf("update question: %d %s", w.Code, w.Body)
	}
	if w := req(otherToken, http.MethodDelete, "/admin/questions/"+itoa(q.ID), ""); w.Code != http.StatusNotFound {
		t.Errorf("unrelated author deletes question: %d", w.Code)
	}

	task := `{"title":"Expose","kind":"shell","format":"md","difficulty":"medium","sandbox_image":"golearn/sandbox-k8s:latest","check_script":"kubectl get svc web","glossary":[{"term":"svc","definition":"Service"}]}`
	if w := req(coToken, http.MethodPost, path+"/tasks", strings.Replace(task, "golearn/sandbox-k8s:latest", "evil/image", 1)); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("unknown sandbox image: %d", w.Code)
	}
	w = req(coToken, http.MethodPost, path+"/tasks", task)
	tk := decode[apigen.AdminTask](t, w)
	if w.Code != http.StatusCreated || tk.LessonID != created.ID || len(tk.Glossary) != 1 || tk.TestCases == nil {
		t.Fatalf("create task: %d %+v", w.Code, tk)
	}
	if w := req(coToken, http.MethodPut, "/admin/tasks/"+itoa(tk.ID), strings.Replace(task, "Expose", "Expose web", 1)); w.Code != http.StatusOK || decode[apigen.AdminTask](t, w).CheckScript != "kubectl get svc web" {
		t.Errorf("update task: %d %s", w.Code, w.Body)
	}

	detail := decode[apigen.AdminLessonDetail](t, req(coToken, http.MethodGet, path, ""))
	if detail.Lesson.Content != "# S" || len(detail.Questions) != 1 || len(detail.Tasks) != 1 || detail.Questions[0].CorrectIndex != 1 {
		t.Errorf("lesson detail = %+v", detail)
	}
	rows := decode[apigen.AdminCourseDetail](t, req(coToken, http.MethodGet, "/admin/courses/12", "")).Lessons
	if len(rows) != 2 || rows[1].QuestionsCount != 1 || rows[1].TasksCount != 1 {
		t.Errorf("course lessons = %+v", rows)
	}

	update := strings.Replace(lesson, `"title":"Services"`, `"title":"Services 2"`, 1)
	if w := req(coToken, http.MethodPut, path, update); w.Code != http.StatusOK {
		t.Errorf("update lesson: %d %s", w.Code, w.Body)
	} else if got := decode[apigen.AdminLesson](t, w); got.Title != "Services 2" || got.OrderNum != 2 || got.Published {
		t.Errorf("updated lesson = %+v", got)
	}
	if w := req(coToken, http.MethodPut, path, strings.Replace(lesson, `"services"`, `"pods"`, 1)); w.Code != http.StatusConflict {
		t.Errorf("rename to taken slug: %d", w.Code)
	}
	if w := req(coToken, http.MethodPut, path+"/published", `{"published":true}`); w.Code != http.StatusForbidden {
		t.Errorf("co-author publishes lesson: %d", w.Code)
	}
	if w := req(coToken, http.MethodPost, path+"/duplicate", ""); w.Code != http.StatusCreated || decode[apigen.AdminLesson](t, w).Published {
		t.Errorf("duplicate: %d %s", w.Code, w.Body)
	}
	if w := req(coToken, http.MethodPost, path+"/move", `{"direction":"up"}`); w.Code != http.StatusNoContent {
		t.Errorf("move: %d", w.Code)
	}

	if w := req(coToken, http.MethodDelete, "/admin/lessons/120", ""); w.Code != http.StatusForbidden {
		t.Errorf("co-author deletes published lesson: %d", w.Code)
	}
	if w := req(coToken, http.MethodDelete, "/admin/tasks/"+itoa(tk.ID), ""); w.Code != http.StatusNoContent || len(c.tasks[created.ID]) != 0 {
		t.Errorf("delete task: %d", w.Code)
	}
	if w := req(coToken, http.MethodDelete, "/admin/questions/"+itoa(q.ID), ""); w.Code != http.StatusNoContent || len(c.questions[created.ID]) != 0 {
		t.Errorf("delete question: %d", w.Code)
	}
	if w := req(coToken, http.MethodDelete, path, ""); w.Code != http.StatusNoContent {
		t.Errorf("co-author deletes draft lesson: %d", w.Code)
	}
	if w := req(ownerToken, http.MethodDelete, "/admin/lessons/120", ""); w.Code != http.StatusNoContent {
		t.Errorf("owner deletes published lesson: %d", w.Code)
	}
}

func TestAdminLessonRoundTripKeepsFields(t *testing.T) {
	h, c := authorsFixture(t)
	c.lessons = append(c.lessons, model.Lesson{ID: 121, ModuleID: 12, Slug: "cfg", Title: "Config", Kind: "lab", Format: "md", Difficulty: "beginner", VMInit: "k8s-start", OrderNum: 2})
	c.questions[121] = []model.QuizQuestion{{ID: 7001, QuizID: 121, Question: "Q", Options: []string{"a", "b"}, OptionExpl: []string{"x", "y"}, CorrectIndex: 1}}
	c.tasks[121] = []model.Task{{ID: 8001, LessonID: 121, Title: "T", Kind: "shell", Format: "md", Difficulty: "easy",
		Glossary: []model.GlossaryItem{{Term: "a", Definition: "b"}}, TestCases: []model.TestCase{{Input: "1", ExpectedOutput: "2"}}}}

	detail := decode[apigen.AdminLessonDetail](t, do(h, http.MethodGet, "/admin/lessons/121", "", withCookie(coToken)))
	for _, step := range []struct{ path, body string }{
		{"/admin/lessons/121", mustJSON(t, detail.Lesson)},
		{"/admin/questions/7001", mustJSON(t, detail.Questions[0])},
		{"/admin/tasks/8001", mustJSON(t, detail.Tasks[0])},
	} {
		if w := do(h, http.MethodPut, step.path, step.body, withCookie(coToken)); w.Code != http.StatusOK {
			t.Fatalf("PUT %s: %d %s", step.path, w.Code, w.Body)
		}
	}
	if l := c.lessons[len(c.lessons)-1]; l.VMInit != "k8s-start" {
		t.Errorf("vm_init lost: %+v", l)
	}
	if q := c.questions[121][0]; len(q.OptionExpl) != 2 {
		t.Errorf("option explanations lost: %+v", q)
	}
	if tk := c.tasks[121][0]; len(tk.Glossary) != 1 || len(tk.TestCases) != 1 {
		t.Errorf("glossary or test cases lost: %+v", tk)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
