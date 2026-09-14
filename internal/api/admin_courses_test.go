package api

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/repository"
)

const ownerToken, coToken, otherToken = "owner-token", "co-token", "other-token"

// authorsFixture adds three authors to the storefront: the owner and a co-author of k8s-draft, and an unrelated author.
func authorsFixture(t *testing.T) (http.Handler, *fakeContent) {
	t.Helper()
	users, c := storefront()
	for i, u := range []*repository.User{
		{ID: 3, Email: "owner@example.com", Name: "Owner", Role: "author"},
		{ID: 4, Email: "co@example.com", Name: "Co", Role: "author"},
		{ID: 5, Email: "other@example.com", Name: "Other", Role: "author"},
	} {
		users.byEmail[u.Email] = u
		users.sessions[[]string{ownerToken, coToken, otherToken}[i]] = u.ID
	}
	owner := 3
	c.modules[2].OwnerID = &owner
	c.coauthors[12] = []int{4}
	return newTestAPIWith(t, users, c), c
}

func TestAdminRoles(t *testing.T) {
	h, _ := authorsFixture(t)
	if w := do(h, http.MethodGet, "/admin/courses", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d", w.Code)
	}
	if w := do(h, http.MethodGet, "/admin/courses", "", withCookie(studentToken)); w.Code != http.StatusForbidden || errorCode(t, w) != codeForbidden {
		t.Errorf("student: %d %s", w.Code, w.Body)
	}
	tests := []struct {
		token     string
		wantSlugs string
		level     apigen.CourseAccessLevel
		publish   bool
	}{
		{adminToken, "linux,docker,k8s-draft,helm,pentest,gym-linux", apigen.CourseAccessLevelAdmin, true},
		{ownerToken, "k8s-draft", apigen.CourseAccessLevelOwner, true},
		{coToken, "k8s-draft", apigen.CourseAccessLevelCoauthor, false},
		{otherToken, "", "", false},
	}
	for _, tt := range tests {
		w := do(h, http.MethodGet, "/admin/courses", "", withCookie(tt.token))
		rows := decode[[]apigen.AdminCourseRow](t, w)
		var slugs []string
		for _, r := range rows {
			slugs = append(slugs, r.Slug)
			if r.Access.Level != tt.level || r.Access.CanPublish != tt.publish || r.Access.CanReorder != (tt.level == apigen.CourseAccessLevelAdmin) {
				t.Errorf("%s: %s access = %+v", tt.token, r.Slug, r.Access)
			}
		}
		if got := strings.Join(slugs, ","); got != tt.wantSlugs || !strings.HasPrefix(w.Body.String(), "[") {
			t.Errorf("%s: courses = %q (%s)", tt.token, got, w.Body)
		}
	}
	rows := decode[[]apigen.AdminCourseRow](t, do(h, http.MethodGet, "/admin/courses", "", withCookie(ownerToken)))
	if rows[0].Owner == nil || rows[0].Owner.ID != 3 || rows[0].LessonsCount != 1 || rows[0].Tags == nil {
		t.Errorf("row = %+v", rows[0])
	}
}

func TestAdminCourseAccess(t *testing.T) {
	h, c := authorsFixture(t)
	course := `{"slug":"k8s","title":"Kubernetes","track":"devops","difficulty":"advanced","label":"challenge","tags":[" pods ",""]}`
	tests := []struct {
		name, token, method, path, body string
		want                            int
	}{
		{"unrelated author cannot see", otherToken, http.MethodGet, "/admin/courses/12", "", http.StatusNotFound},
		{"author cannot see platform course", ownerToken, http.MethodGet, "/admin/courses/10", "", http.StatusNotFound},
		{"admin sees platform course", adminToken, http.MethodGet, "/admin/courses/10", "", http.StatusOK},
		{"co-author reads", coToken, http.MethodGet, "/admin/courses/12", "", http.StatusOK},
		{"co-author cannot publish", coToken, http.MethodPut, "/admin/courses/12/published", `{"published":true}`, http.StatusForbidden},
		{"co-author cannot publish via update", coToken, http.MethodPut, "/admin/courses/12", strings.Replace(course, `"tags"`, `"published":true,"tags"`, 1), http.StatusForbidden},
		{"co-author edits", coToken, http.MethodPut, "/admin/courses/12", course, http.StatusOK},
		{"co-author cannot delete", coToken, http.MethodDelete, "/admin/courses/12", "", http.StatusForbidden},
		{"owner cannot reorder", ownerToken, http.MethodPost, "/admin/courses/12/move", `{"direction":"up"}`, http.StatusForbidden},
		{"admin reorders", adminToken, http.MethodPost, "/admin/courses/12/move", `{"direction":"up"}`, http.StatusNoContent},
		{"bad direction", adminToken, http.MethodPost, "/admin/courses/12/move", `{"direction":"left"}`, http.StatusUnprocessableEntity},
		{"owner publishes", ownerToken, http.MethodPut, "/admin/courses/12/published", `{"published":true}`, http.StatusNoContent},
		{"non-numeric id", adminToken, http.MethodGet, "/admin/courses/abc", "", http.StatusNotFound},
		{"owner deletes", ownerToken, http.MethodDelete, "/admin/courses/12", "", http.StatusNoContent},
	}
	for _, tt := range tests {
		w := do(h, tt.method, tt.path, tt.body, withCookie(tt.token))
		if w.Code != tt.want {
			t.Fatalf("%s: status %d, want %d (%s)", tt.name, w.Code, tt.want, w.Body)
		}
		if tt.name == "co-author edits" {
			got := decode[apigen.AdminCourse](t, w)
			m := c.modules[2]
			if got.Slug != "k8s" || m.Published || m.OwnerID == nil || *m.OwnerID != 3 || m.OrderNum != 3 || strings.Join(m.Tags, ",") != "pods" || got.Label == nil || *got.Label != "challenge" {
				t.Errorf("updated = %+v, stored %+v", got, m)
			}
		}
	}
	if len(c.modules) != 5 {
		t.Errorf("course not deleted: %d", len(c.modules))
	}
}

func TestAdminCreateCourse(t *testing.T) {
	h, c := authorsFixture(t)
	post := func(token, body string) *httptest.ResponseRecorder {
		return do(h, http.MethodPost, "/admin/courses", body, withCookie(token))
	}
	w := post(otherToken, `{"slug":"ansible","title":" Ansible ","track":"devops","difficulty":"beginner","cover_url":"https://cdn.example.com/a.png"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	got := decode[apigen.AdminCourse](t, w)
	if got.Title != "Ansible" || got.Published || got.Owner == nil || got.Owner.ID != 5 || got.Access.Level != apigen.CourseAccessLevelOwner ||
		got.OrderNum != 5 || got.CoverURL == nil || !got.HasCustomCover || got.Source != "admin" || got.Label != nil {
		t.Errorf("created = %+v", got)
	}
	if w := post(otherToken, `{"slug":"ansible","title":"Dup","track":"devops","difficulty":"beginner"}`); w.Code != http.StatusConflict || errorCode(t, w) != codeSlugTaken {
		t.Errorf("duplicate slug: %d %s", w.Code, w.Body)
	}
	if w := post(otherToken, `{"slug":"gymx","title":"Gym","track":"gym","difficulty":"beginner","published":true}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("author gym track: %d %s", w.Code, w.Body)
	}
	if w := post(adminToken, `{"slug":"gymx","title":"Gym","track":"gym","difficulty":"beginner","published":true}`); w.Code != http.StatusCreated || !decode[apigen.AdminCourse](t, w).Published {
		t.Errorf("admin gym track: %d %s", w.Code, w.Body)
	}

	w = post(otherToken, `{"slug":"Bad Slug","title":"","track":"nope","difficulty":"easy","label":"hard","est_minutes":-1,"cover_url":"javascript:alert(1)"}`)
	e := decode[apigen.Error](t, w)
	want := map[string]string{"slug": fieldInvalidFormat, "title": fieldRequired, "track": fieldNotFound, "difficulty": fieldInvalidValue,
		"label": fieldInvalidValue, "est_minutes": fieldInvalidValue, "cover_url": fieldInvalidFormat}
	if w.Code != http.StatusUnprocessableEntity || e.Error.Details == nil || e.Error.Details.Fields == nil {
		t.Fatalf("validation: %d %s", w.Code, w.Body)
	}
	for k, v := range want {
		if (*e.Error.Details.Fields)[k] != v {
			t.Errorf("field %s = %q, want %q", k, (*e.Error.Details.Fields)[k], v)
		}
	}
	if w := post(studentToken, `{}`); w.Code != http.StatusForbidden {
		t.Errorf("student create: %d", w.Code)
	}
	if n := len(c.modules); n != 8 {
		t.Errorf("modules = %d", n)
	}
}

func TestDraftPreviewForEditors(t *testing.T) {
	h, _ := authorsFixture(t)
	for _, tt := range []struct {
		token string
		want  int
	}{{coToken, http.StatusOK}, {ownerToken, http.StatusOK}, {adminToken, http.StatusOK}, {otherToken, http.StatusNotFound}, {studentToken, http.StatusNotFound}} {
		for _, path := range []string{"/courses/k8s-draft", "/courses/k8s-draft/lessons/pods"} {
			if w := do(h, http.MethodGet, path, "", withCookie(tt.token)); w.Code != tt.want {
				t.Errorf("%s %s: %d, want %d", tt.token, path, w.Code, tt.want)
			}
		}
	}
}

func TestCourseAuthors(t *testing.T) {
	h, c := authorsFixture(t)
	add := func(token, email string) *httptest.ResponseRecorder {
		return do(h, http.MethodPost, "/admin/courses/12/authors", `{"email":"`+email+`"}`, withCookie(token))
	}
	fieldOf := func(w *httptest.ResponseRecorder, name string) string {
		e := decode[apigen.Error](t, w)
		if e.Error.Details == nil || e.Error.Details.Fields == nil {
			return ""
		}
		return (*e.Error.Details.Fields)[name]
	}
	if w := add(ownerToken, "s@example.com"); w.Code != http.StatusUnprocessableEntity || fieldOf(w, "email") != fieldNotAuthor {
		t.Errorf("student as co-author: %d %s", w.Code, w.Body)
	}
	if w := add(ownerToken, "ghost@example.com"); w.Code != http.StatusUnprocessableEntity || fieldOf(w, "email") != fieldNotFound {
		t.Errorf("unknown email: %d %s", w.Code, w.Body)
	}
	if w := add(coToken, "other@example.com"); w.Code != http.StatusForbidden {
		t.Errorf("co-author adds: %d", w.Code)
	}
	if w := add(ownerToken, " OTHER@example.com "); w.Code != http.StatusCreated || decode[apigen.AuthorRef](t, w).ID != 5 {
		t.Errorf("owner adds: %d %s", w.Code, w.Body)
	}
	for _, email := range []string{"other@example.com", "owner@example.com"} {
		if w := add(ownerToken, email); w.Code != http.StatusConflict || errorCode(t, w) != codeAlreadyAuthor {
			t.Errorf("re-add %s: %d %s", email, w.Code, w.Body)
		}
	}
	if w := do(h, http.MethodDelete, "/admin/courses/12/authors/5", "", withCookie(coToken)); w.Code != http.StatusForbidden {
		t.Errorf("co-author removes another: %d", w.Code)
	}
	if w := do(h, http.MethodDelete, "/admin/courses/12/authors/4", "", withCookie(coToken)); w.Code != http.StatusNoContent {
		t.Errorf("co-author leaves: %d %s", w.Code, w.Body)
	}
	if w := do(h, http.MethodGet, "/admin/courses/12", "", withCookie(coToken)); w.Code != http.StatusNotFound {
		t.Errorf("after leaving: %d", w.Code)
	}
	if w := do(h, http.MethodDelete, "/admin/courses/12/authors/4", "", withCookie(ownerToken)); w.Code != http.StatusNotFound {
		t.Errorf("remove non-author: %d", w.Code)
	}

	if w := do(h, http.MethodPut, "/admin/courses/12/owner", `{"user_id":5}`, withCookie(ownerToken)); w.Code != http.StatusForbidden {
		t.Errorf("owner transfers: %d", w.Code)
	}
	if w := do(h, http.MethodPut, "/admin/courses/12/owner", `{"user_id":1}`, withCookie(adminToken)); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("transfer to student: %d", w.Code)
	}
	w := do(h, http.MethodPut, "/admin/courses/12/owner", `{"user_id":5}`, withCookie(adminToken))
	got := decode[apigen.CourseAuthors](t, w)
	if w.Code != http.StatusOK || got.Owner == nil || got.Owner.ID != 5 || len(got.Coauthors) != 1 || got.Coauthors[0].ID != 3 || *c.modules[2].OwnerID != 5 {
		t.Errorf("transfer: %d %+v", w.Code, got)
	}
	if w := do(h, http.MethodDelete, "/admin/courses/12", "", withCookie(ownerToken)); w.Code != http.StatusForbidden {
		t.Errorf("former owner deletes: %d", w.Code)
	}
}

func TestCourseCover(t *testing.T) {
	h, c := authorsFixture(t)
	upload := func(token, name string, data []byte) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, _ := mw.CreateFormFile(name, "cover.bin")
		_, _ = fw.Write(data)
		_ = mw.Close()
		return do(h, http.MethodPut, "/admin/courses/12/cover", "", withCookie(token), withBody(buf.Bytes(), mw.FormDataContentType()))
	}
	png := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...)
	if w := upload(coToken, "file", png); w.Code != http.StatusNoContent {
		t.Fatalf("upload png: %d %s", w.Code, w.Body)
	}
	if !strings.HasPrefix(c.modules[2].CoverImage, "data:image/png;base64,") {
		t.Errorf("stored cover = %.40q", c.modules[2].CoverImage)
	}
	w := do(h, http.MethodGet, "/courses/k8s-draft/cover", "")
	if w.Header().Get("Content-Type") != "image/png" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("served cover headers = %v", w.Header())
	}
	course := decode[apigen.AdminCourseDetail](t, do(h, http.MethodGet, "/admin/courses/12", "", withCookie(coToken))).Course
	if !course.HasCustomCover || course.CoverURL != nil {
		t.Errorf("course cover fields = %+v", course)
	}

	if w := upload(coToken, "file", []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)); w.Code != http.StatusNoContent || !strings.HasPrefix(c.modules[2].CoverImage, "data:image/svg+xml;") {
		t.Errorf("upload svg: %d %s", w.Code, w.Body)
	}
	tests := []struct {
		name, field string
		data        []byte
		want        int
	}{
		{"text file", "file", []byte("hello"), http.StatusUnsupportedMediaType},
		{"missing file part", "other", png, http.StatusUnprocessableEntity},
		{"too large", "file", append(png, make([]byte, maxCoverBytes)...), http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		if w := upload(coToken, tt.field, tt.data); w.Code != tt.want {
			t.Errorf("%s: %d, want %d (%s)", tt.name, w.Code, tt.want, w.Body)
		}
	}
	if w := do(h, http.MethodPut, "/admin/courses/12/cover", `{}`, withCookie(coToken)); w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("json body: %d", w.Code)
	}
	if w := upload(otherToken, "file", png); w.Code != http.StatusNotFound {
		t.Errorf("unrelated author upload: %d", w.Code)
	}
	if w := do(h, http.MethodDelete, "/admin/courses/12/cover", "", withCookie(coToken)); w.Code != http.StatusNoContent || c.modules[2].CoverImage != "" {
		t.Errorf("delete cover: %d", w.Code)
	}
}

func TestPreviewContent(t *testing.T) {
	h, _ := authorsFixture(t)
	body := `{"format":"md","content":"# Title\n\n<script>alert(1)</script>"}`
	if w := do(h, http.MethodPost, "/admin/content/preview", body, withCookie(studentToken)); w.Code != http.StatusForbidden {
		t.Errorf("student: %d", w.Code)
	}
	w := do(h, http.MethodPost, "/admin/content/preview", body, withCookie(otherToken))
	out := decode[map[string]string](t, w)
	if w.Code != http.StatusOK || !strings.Contains(out["html"], "<h1") || strings.Contains(out["html"], "<script") {
		t.Errorf("preview: %d %q", w.Code, out["html"])
	}
	if w := do(h, http.MethodPost, "/admin/content/preview", `{"format":"rst","content":""}`, withCookie(otherToken)); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("bad format: %d", w.Code)
	}
}

func withBody(data []byte, contentType string) reqOpt {
	return func(r *http.Request) {
		r.Body = io.NopCloser(bytes.NewReader(data))
		r.ContentLength = int64(len(data))
		r.Header.Set("Content-Type", contentType)
	}
}
