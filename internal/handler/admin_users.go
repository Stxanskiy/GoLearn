package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/backendraz/golearn/internal/repository"
	"github.com/go-chi/chi/v5"
)

type AdminUsersData struct {
	PageTitle string
	Users     []repository.User
	Error     string
	Created   string
}

// AdminUsers renders the admin user-management page: a create form + the list
// of existing users. Guarded by AdminMiddleware.
func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.List(r.Context())
	if err != nil {
		h.log.Error("admin list users", "error", err)
	}
	q := r.URL.Query()
	h.render(w, "admin_users", &AdminUsersData{
		PageTitle: "Пользователи",
		Users:     users,
		Error:     q.Get("err"),
		Created:   q.Get("created"),
	})
}

// AdminUserCreate handles POST /admin/users — creates a user with the chosen role.
// Flash messages are passed back via query params (PRG pattern).
func (h *Handler) AdminUserCreate(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	role := strings.TrimSpace(r.FormValue("role"))
	if role != "admin" {
		role = "student"
	}

	fail := func(msg string) {
		http.Redirect(w, r, "/admin/users?err="+url.QueryEscape(msg), http.StatusSeeOther)
	}
	if name == "" || email == "" || len(password) < 6 {
		fail("Заполните все поля (пароль от 6 символов)")
		return
	}
	if !strings.Contains(email, "@") {
		fail("Некорректный email")
		return
	}

	if _, err := h.userRepo.CreateWithRole(r.Context(), email, password, name, role); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			fail("Email уже зарегистрирован")
			return
		}
		h.log.Error("admin create user", "error", err)
		fail("Ошибка создания пользователя")
		return
	}
	http.Redirect(w, r, "/admin/users?created="+url.QueryEscape(email), http.StatusSeeOther)
}

// selfID returns the current admin's id (0 if unknown) — used to block self-harm.
func (h *Handler) selfID(r *http.Request) int {
	if u := GetUser(r.Context()); u != nil {
		return u.ID
	}
	return 0
}

func usersBack(w http.ResponseWriter, r *http.Request, msg string) {
	u := "/admin/users"
	if msg != "" {
		u += "?err=" + url.QueryEscape(msg)
	}
	http.Redirect(w, r, u, http.StatusSeeOther)
}

// AdminUserSetRole changes a user's role. An admin cannot change their own role
// (avoids locking themselves out of the admin panel).
func (h *Handler) AdminUserSetRole(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	if id == h.selfID(r) {
		usersBack(w, r, "Нельзя менять собственную роль")
		return
	}
	role := strings.TrimSpace(r.FormValue("role"))
	if role != "admin" {
		role = "student"
	}
	_ = h.userRepo.SetRole(r.Context(), id, role)
	usersBack(w, r, "")
}

// AdminUserBlock blocks/unblocks a user (blocked=1 blocks). Self-block is refused.
func (h *Handler) AdminUserBlock(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	if id == h.selfID(r) {
		usersBack(w, r, "Нельзя заблокировать себя")
		return
	}
	_ = h.userRepo.SetBlocked(r.Context(), id, r.FormValue("blocked") == "1")
	usersBack(w, r, "")
}

// AdminUserDelete removes a user. Self-delete is refused.
func (h *Handler) AdminUserDelete(w http.ResponseWriter, r *http.Request) {
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	if id == h.selfID(r) {
		usersBack(w, r, "Нельзя удалить себя")
		return
	}
	_ = h.userRepo.Delete(r.Context(), id)
	usersBack(w, r, "")
}

// AdminUserResetPassword sets a new password for a user (admin reset).
func (h *Handler) AdminUserResetPassword(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	pw := r.FormValue("password")
	if len(pw) < 6 {
		usersBack(w, r, "Пароль от 6 символов")
		return
	}
	if err := h.userRepo.SetPassword(r.Context(), id, pw); err != nil {
		usersBack(w, r, "Ошибка сброса пароля")
		return
	}
	usersBack(w, r, "")
}
