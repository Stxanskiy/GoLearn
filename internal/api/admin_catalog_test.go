package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
)

func TestAdminSpecializations(t *testing.T) {
	h, c := authorsFixture(t)
	req := func(token, method, path, body string) (int, string) {
		w := do(h, method, path, body, withCookie(token))
		return w.Code, w.Body.String()
	}

	w := do(h, http.MethodGet, "/admin/specializations", "", withCookie(otherToken))
	specs := decode[[]apigen.AdminSpecialization](t, w)
	counts := map[string]int{}
	for _, s := range specs {
		counts[s.Slug] = s.CoursesCount
	}
	if w.Code != http.StatusOK || counts["devops"] != 4 || counts["security"] != 1 || counts["gym"] != 1 || counts["database"] != 0 {
		t.Errorf("course counts = %v", counts)
	}
	for _, tt := range []struct{ method, path, body string }{
		{http.MethodPost, "/admin/specializations", `{"slug":"go","name":"Go"}`},
		{http.MethodPut, "/admin/specializations/devops", `{"name":"X"}`},
		{http.MethodDelete, "/admin/specializations/database", ""},
		{http.MethodGet, "/admin/simulators", ""},
	} {
		if code, _ := req(otherToken, tt.method, tt.path, tt.body); code != http.StatusForbidden {
			t.Errorf("author %s %s: %d", tt.method, tt.path, code)
		}
	}

	if code, body := req(adminToken, http.MethodPost, "/admin/specializations", `{"slug":"frontend","name":"Frontend","icon":"🖥","published":true}`); code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, body)
	}
	if s := c.specs[len(c.specs)-1]; s.Slug != "frontend" || !s.Published || s.OwnerID == nil || *s.OwnerID != 2 || s.OrderNum != 5 {
		t.Errorf("created = %+v", s)
	}
	if code, _ := req(adminToken, http.MethodPost, "/admin/specializations", `{"slug":"frontend","name":"Dup"}`); code != http.StatusConflict {
		t.Errorf("duplicate: %d", code)
	}
	code, body := req(adminToken, http.MethodPost, "/admin/specializations", `{"slug":"a-very-long-specialization-slug-over-forty-chars","name":""}`)
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, `"slug":"too_long"`) || !strings.Contains(body, `"name":"required"`) {
		t.Errorf("validation: %d %s", code, body)
	}

	c.specs[len(c.specs)-1].CoverImage = "data:image/png;base64,AAAA"
	code, body = req(adminToken, http.MethodPut, "/admin/specializations/frontend", `{"name":"Front","description":"UI"}`)
	if s := c.specs[len(c.specs)-1]; code != http.StatusOK || s.Name != "Front" || !s.Published || s.CoverImage == "" || s.OrderNum != 5 {
		t.Errorf("update: %d %s stored %+v", code, body, s)
	}

	if code, body = req(otherToken, http.MethodPost, "/admin/courses", `{"slug":"react","title":"React","track":"frontend","difficulty":"beginner"}`); code != http.StatusCreated {
		t.Fatalf("course in new specialization: %d %s", code, body)
	}
	if code, body = req(adminToken, http.MethodDelete, "/admin/specializations/frontend", ""); code != http.StatusConflict || !strings.Contains(body, codeSpecNotEmpty) {
		t.Errorf("delete non-empty: %d %s", code, body)
	}
	if code, _ = req(adminToken, http.MethodDelete, "/admin/specializations/database", ""); code != http.StatusNoContent || (fakeSpecs{c}).spec("database") != nil {
		t.Errorf("delete empty: %d", code)
	}
	if code, _ = req(adminToken, http.MethodPut, "/admin/specializations/nope/published", `{"published":true}`); code != http.StatusNotFound {
		t.Errorf("unknown specialization: %d", code)
	}
	if code, _ = req(adminToken, http.MethodDelete, "/admin/specializations/frontend/cover", ""); code != http.StatusNoContent || (fakeSpecs{c}).spec("frontend").CoverImage != "" {
		t.Errorf("delete cover: %d", code)
	}
}

func TestAdminSimulators(t *testing.T) {
	h, c := authorsFixture(t)
	c.sims = []model.Simulator{{Slug: "sre", Title: "SRE", OrderNum: 1, Published: true, Data: `{"slug":"sre","title":"SRE","metrics":[],"turns":[{"title":"t","situation":"s","choices":[{"text":"a","effects":{},"result":"r"}]}]}`}}
	req := func(method, path, body string) (int, string) {
		w := do(h, method, path, body, withCookie(adminToken))
		return w.Code, w.Body.String()
	}
	scenario := `{"slug":"junior","title":"Junior","role":"DevOps","icon":"🧑","intro":"Hi",
		"metrics":[{"key":"budget","label":"Budget","unit":"$","start":100,"higher":true}],
		"turns":[{"title":"Day 1","situation":"Outage","choices":[{"text":"Fix","effects":{"budget":-10},"result":"ok"}]}]}`

	if code, body := req(http.MethodPost, "/admin/simulators", `{"scenario":`+strings.Replace(scenario, `"budget":-10`, `"morale":1`, 1)+`}`); code != http.StatusUnprocessableEntity || !strings.Contains(body, "scenario.turns") {
		t.Errorf("effect on unknown metric: %d %s", code, body)
	}
	code, body := req(http.MethodPost, "/admin/simulators", `{"scenario":`+scenario+`}`)
	if code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, body)
	}
	created := decode[apigen.AdminSimulator](t, do(h, http.MethodGet, "/admin/simulators/junior", "", withCookie(adminToken)))
	if created.Published || created.OrderNum != 2 || created.Scenario.Turns[0].Choices[0].Effects["budget"] != -10 || created.OwnerID == nil {
		t.Errorf("created = %+v", created)
	}
	if code, _ := req(http.MethodPost, "/admin/simulators", `{"scenario":`+scenario+`}`); code != http.StatusConflict {
		t.Errorf("duplicate: %d", code)
	}
	if code, body := req(http.MethodPut, "/admin/simulators/sre", `{"scenario":`+scenario+`}`); code != http.StatusUnprocessableEntity || !strings.Contains(body, "scenario.slug") {
		t.Errorf("slug mismatch: %d %s", code, body)
	}
	if code, body := req(http.MethodPut, "/admin/simulators/junior", `{"published":true,"scenario":`+strings.Replace(scenario, `"Junior"`, `"Junior 2"`, 1)+`}`); code != http.StatusOK || decode[apigen.AdminSimulator](t, do(h, http.MethodGet, "/admin/simulators/junior", "", withCookie(adminToken))).Scenario.Title != "Junior 2" {
		t.Errorf("update: %d %s", code, body)
	}
	if w := do(h, http.MethodGet, "/simulators/junior", "", withCookie(studentToken)); w.Code != http.StatusOK {
		t.Errorf("published simulator for students: %d", w.Code)
	}
	rows := decode[[]apigen.AdminSimulatorRow](t, do(h, http.MethodGet, "/admin/simulators", "", withCookie(adminToken)))
	if len(rows) != 2 || rows[1].Title != "Junior 2" {
		t.Errorf("rows = %+v", rows)
	}
	if code, _ := req(http.MethodPut, "/admin/simulators/junior/published", `{"published":false}`); code != http.StatusNoContent || c.sims[1].Published {
		t.Errorf("unpublish: %d", code)
	}
	if code, _ := req(http.MethodPost, "/admin/simulators/junior/move", `{"direction":"up"}`); code != http.StatusNoContent {
		t.Errorf("move: %d", code)
	}
	if code, _ := req(http.MethodDelete, "/admin/simulators/junior", ""); code != http.StatusNoContent || len(c.sims) != 1 {
		t.Errorf("delete: %d", code)
	}
}
