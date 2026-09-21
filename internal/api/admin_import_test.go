package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
)

func TestExportCourse(t *testing.T) {
	h, c := authorsFixture(t)
	c.tasks[120] = []model.Task{{ID: 901, LessonID: 120, Title: "Run pod", Kind: "shell", CheckScript: "kubectl get pod web"}}
	w := do(h, http.MethodGet, "/admin/courses/12/export", "", withCookie(coToken))
	if w.Code != http.StatusOK || w.Header().Get("Content-Disposition") != `attachment; filename="k8s-draft.course.json"` {
		t.Fatalf("export: %d %v", w.Code, w.Header())
	}
	var doc apigen.CourseDocument
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil || doc.Slug != "k8s-draft" || len(doc.Lessons) != 1 || (*doc.Lessons[0].Tasks)[0].CheckScript == nil {
		t.Errorf("document = %+v (%v)", doc, err)
	}
	if w := do(h, http.MethodGet, "/admin/courses/12/export", "", withCookie(otherToken)); w.Code != http.StatusNotFound {
		t.Errorf("unrelated author export: %d", w.Code)
	}
}

func TestImportCourse(t *testing.T) {
	h, c := authorsFixture(t)
	post := func(token, path, body string) (int, string) {
		w := do(h, http.MethodPost, path, body, withCookie(token))
		return w.Code, w.Body.String()
	}
	doc := `{"slug":"ansible","title":"Ansible","track":"devops","difficulty":"beginner","label":"Старт","order_num":1,
		"lessons":[{"slug":"intro","title":"Intro","kind":"theory","quiz":[{"question":"Q","options":["a","b"],"correct":2}]},
		           {"slug":"lab","title":"Lab","kind":"lab","tasks":[{"title":"Ping","kind":"shell","sandbox_image":"golearn/sandbox:latest","check_script":"true"}]}]}`

	code, body := post(otherToken, "/admin/import/preview", doc)
	var preview apigen.ImportPreview
	_ = json.Unmarshal([]byte(body), &preview)
	if code != http.StatusOK || preview.Exists || len(preview.NewLessons) != 2 || preview.Blocked || preview.Issues == nil {
		t.Fatalf("preview new: %d %s", code, body)
	}
	if n := len(c.modules); n != 6 {
		t.Fatalf("preview wrote modules: %d", n)
	}

	blocked := strings.Replace(doc, "golearn/sandbox:latest", "evil/image", 1)
	code, body = post(otherToken, "/admin/import", blocked)
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, codeImportBlocked) || !strings.Contains(body, "unknown_sandbox_image") || !strings.Contains(body, "/lessons/1/tasks/0/sandbox_image") {
		t.Errorf("blocked import: %d %s", code, body)
	}
	if code, body = post(otherToken, "/admin/import", strings.Replace(doc, `"track":"devops"`, `"track":"nope"`, 1)); code != http.StatusUnprocessableEntity || !strings.Contains(body, `"track":"not_found"`) {
		t.Errorf("author imports into an unknown track: %d %s", code, body)
	}
	if code, body = post(otherToken, "/admin/import", strings.Replace(doc, `"label":"Старт"`, `"cover_image":"javascript:alert(1)"`, 1)); code != http.StatusUnprocessableEntity || !strings.Contains(body, "cover_image") {
		t.Errorf("bad cover: %d %s", code, body)
	}

	code, body = post(otherToken, "/admin/import", doc)
	var res apigen.ImportResult
	_ = json.Unmarshal([]byte(body), &res)
	if code != http.StatusOK || !res.Created {
		t.Fatalf("import new: %d %s", code, body)
	}
	m := fakeModules{c}.module(res.CourseID)
	if m == nil || m.Published || m.OwnerID == nil || *m.OwnerID != 5 || m.OrderNum != 5 || m.Label != "start" {
		t.Errorf("imported course = %+v", m)
	}
	for _, l := range c.lessons {
		if l.ModuleID == res.CourseID && !l.Published {
			t.Errorf("lesson of a new draft course should be published: %+v", l)
		}
	}

	if code, _ := post(otherToken, "/admin/import/preview", strings.Replace(doc, `"ansible"`, `"linux"`, 1)); code != http.StatusConflict {
		t.Errorf("import over a platform course by author: %d", code)
	}
	if code, _ := post(otherToken, "/admin/import", strings.Replace(doc, `"ansible"`, `"k8s-draft"`, 1)); code != http.StatusConflict {
		t.Errorf("import over another author's course: %d", code)
	}

	replace := `{"slug":"k8s-draft","title":"Kubernetes","track":"devops","difficulty":"advanced","order_num":1,"lessons":[{"slug":"svc","title":"Services","kind":"theory"}]}`
	code, body = post(coToken, "/admin/import/preview", replace)
	_ = json.Unmarshal([]byte(body), &preview)
	if code != http.StatusOK || !preview.Exists || len(preview.RemovedLessons) != 1 || preview.RemovedLessons[0].Slug != "pods" {
		t.Errorf("co-author preview: %d %s", code, body)
	}
	if code, body = post(coToken, "/admin/import", replace); code != http.StatusForbidden {
		t.Errorf("co-author removes published lesson via import: %d %s", code, body)
	}
	keep := strings.Replace(replace, `[{"slug":"svc"`, `[{"slug":"pods","title":"Pods","kind":"theory"},{"slug":"svc"`, 1)
	if code, body = post(coToken, "/admin/import", keep); code != http.StatusOK {
		t.Fatalf("co-author import: %d %s", code, body)
	}
	k8s := c.modules[2]
	if k8s.Published || k8s.OrderNum != 3 || k8s.OwnerID == nil || *k8s.OwnerID != 3 {
		t.Errorf("existing course state changed: %+v", k8s)
	}
	for _, l := range c.lessons {
		if l.ModuleID == 12 && l.Slug == "svc" && l.Published {
			t.Errorf("co-author's new lesson should be a draft: %+v", l)
		}
	}
	if code, body = post(ownerToken, "/admin/import", replace); code != http.StatusOK {
		t.Errorf("owner removes lesson via import: %d %s", code, body)
	}
}
