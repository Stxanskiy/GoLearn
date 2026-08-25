package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/backendraz/golearn/internal/repository"
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
