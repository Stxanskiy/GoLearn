package api

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/repository"
)

func TestAdminUsers(t *testing.T) {
	t.Setenv("ADMIN_EMAILS", "root@example.com")
	users, c := storefront()
	users.byEmail["root@example.com"] = &repository.User{ID: 6, Email: "root@example.com", Name: "Root", Role: "admin"}
	users.byEmail["bob@example.com"] = &repository.User{ID: 7, Email: "bob@example.com", Name: "Bob", Role: "student"}
	users.sessions["bob-token"] = 7
	h := newTestAPIWith(t, users, c)
	req := func(method, path, body string) (int, string) {
		w := do(h, method, path, body, withCookie(adminToken))
		return w.Code, w.Body.String()
	}

	if w := do(h, http.MethodGet, "/admin/users", "", withCookie(studentToken)); w.Code != http.StatusForbidden {
		t.Errorf("student lists users: %d", w.Code)
	}
	page := decode[apigen.AdminUserPage](t, do(h, http.MethodGet, "/admin/users?limit=2", "", withCookie(adminToken)))
	if len(page.Items) != 2 || page.Items[0].ID != 7 || page.NextCursor == nil || *page.NextCursor != "6" || !page.Items[1].RolePinnedByEnv {
		t.Fatalf("first page = %+v", page)
	}
	page = decode[apigen.AdminUserPage](t, do(h, http.MethodGet, "/admin/users?limit=2&cursor="+*page.NextCursor, "", withCookie(adminToken)))
	if len(page.Items) != 2 || page.Items[0].ID != 2 || !page.Items[0].IsSelf || page.NextCursor != nil {
		t.Errorf("last page = %+v", page)
	}
	if page := decode[apigen.AdminUserPage](t, do(h, http.MethodGet, "/admin/users?q=Bob", "", withCookie(adminToken))); len(page.Items) != 1 {
		t.Errorf("search = %+v", page)
	}
	if code, _ := req(http.MethodGet, "/admin/users?limit=500", ""); code != http.StatusUnprocessableEntity {
		t.Errorf("limit over max: %d", code)
	}

	code, body := req(http.MethodPost, "/admin/users", `{"name":"Ann","email":"ann@example.com","password":"secret1","role":"author"}`)
	if code != http.StatusCreated || decode[apigen.AdminUser](t, do(h, http.MethodPatch, "/admin/users/101", `{"blocked":false}`, withCookie(adminToken))).Role != "author" {
		t.Errorf("create author: %d %s", code, body)
	}
	if code, _ := req(http.MethodPost, "/admin/users", `{"name":"Ann","email":"ann@example.com","password":"secret1","role":"author"}`); code != http.StatusConflict {
		t.Errorf("duplicate email: %d", code)
	}
	if code, body := req(http.MethodPost, "/admin/users", `{"name":"","email":"nope","password":"1","role":"owner"}`); code != http.StatusUnprocessableEntity || !strings.Contains(body, `"role":"invalid_value"`) || !strings.Contains(body, `"email":"invalid_format"`) {
		t.Errorf("validation: %d %s", code, body)
	}

	tests := []struct {
		name, method, path, body string
		want                     int
		code                     string
	}{
		{"change own role", http.MethodPatch, "/admin/users/2", `{"role":"student"}`, http.StatusConflict, codeSelfModification},
		{"delete yourself", http.MethodDelete, "/admin/users/2", ``, http.StatusConflict, codeSelfModification},
		{"demote env admin", http.MethodPatch, "/admin/users/6", `{"role":"author"}`, http.StatusConflict, codeRolePinnedByEnv},
		{"unknown role", http.MethodPatch, "/admin/users/7", `{"role":"owner"}`, http.StatusUnprocessableEntity, codeValidationFailed},
		{"empty patch", http.MethodPatch, "/admin/users/7", `{}`, http.StatusUnprocessableEntity, codeValidationFailed},
		{"block env admin leaves one admin", http.MethodPatch, "/admin/users/6", `{"blocked":true}`, http.StatusOK, ""},
		{"blocked admin can be deleted", http.MethodDelete, "/admin/users/6", ``, http.StatusNoContent, ""},
		{"short password", http.MethodPut, "/admin/users/7/password", `{"password":"123"}`, http.StatusUnprocessableEntity, codeValidationFailed},
		{"missing user", http.MethodDelete, "/admin/users/999", ``, http.StatusNotFound, codeNotFound},
	}
	for _, tt := range tests {
		code, body := req(tt.method, tt.path, tt.body)
		if code != tt.want || (tt.code != "" && !strings.Contains(body, `"code":"`+tt.code+`"`)) {
			t.Errorf("%s: %d %s, want %d %s", tt.name, code, body, tt.want, tt.code)
		}
	}

	bob := users.byEmail["bob@example.com"]
	bob.Role = "admin"
	if code, body := req(http.MethodPatch, "/admin/users/7", `{"role":"author"}`); code != http.StatusOK || !strings.Contains(body, `"role":"author"`) {
		t.Errorf("demote second admin: %d %s", code, body)
	}
	if w := do(h, http.MethodGet, "/admin/users", "", withCookie("bob-token")); w.Code != http.StatusForbidden {
		t.Errorf("demoted admin keeps access: %d", w.Code)
	}
	if code, _ := req(http.MethodPatch, "/admin/users/7", `{"blocked":true}`); code != http.StatusOK {
		t.Errorf("block: %d", code)
	}
	if w := do(h, http.MethodGet, "/me", "", withCookie("bob-token")); w.Code != http.StatusUnauthorized {
		t.Errorf("blocked user session still works: %d", w.Code)
	}
	bob.Blocked = false
	users.sessions["bob-token"] = 7
	if code, _ := req(http.MethodPut, "/admin/users/7/password", `{"password":"newpass1"}`); code != http.StatusNoContent || bob.PasswordHash != "newpass1" {
		t.Errorf("reset password: %d", code)
	}
	if _, ok := users.sessions["bob-token"]; ok {
		t.Errorf("session survived password reset")
	}
}

// lastAdminUsers is a user store whose guarded changes always hit the last active admin.
type lastAdminUsers struct{ *fakeUsers }

func (lastAdminUsers) SetAccess(context.Context, int, string, bool) error {
	return repository.ErrLastAdmin
}

func (lastAdminUsers) DeleteGuarded(context.Context, int) error { return repository.ErrLastAdmin }

func TestLastAdminConflict(t *testing.T) {
	users, c := storefront()
	stores := c.stores(users)
	stores.Users = lastAdminUsers{users}
	h := New(stores, Config{}, slog.New(slog.NewTextHandler(io.Discard, nil))).Routes()
	for _, tt := range []struct{ method, path, body string }{
		{http.MethodPatch, "/admin/users/1", `{"blocked":true}`},
		{http.MethodDelete, "/admin/users/1", ``},
	} {
		w := do(h, tt.method, tt.path, tt.body, withCookie(adminToken))
		if w.Code != http.StatusConflict || errorCode(t, w) != codeLastAdmin {
			t.Errorf("%s %s: %d %s", tt.method, tt.path, w.Code, w.Body)
		}
	}
}
