package handler

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/repository"
)

type contextKey string

const userContextKey contextKey = "user"

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login/register pages
		path := r.URL.Path
		if path == "/login" || path == "/register" || strings.HasPrefix(path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := h.userRepo.GetUserBySession(r.Context(), cookie.Value)
		if err != nil {
			// Invalid/expired session
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "", MaxAge: -1, Path: "/"})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		auth.PromoteEnvAdmin(r.Context(), h.userRepo, user)
		// Non-sensitive hint cookie so the navbar can show the admin link
		// (/admin itself is still guarded by AdminMiddleware).
		if user.IsAdmin() {
			http.SetCookie(w, &http.Cookie{Name: "gl_role", Value: "admin", Path: "/", MaxAge: 30 * 24 * 3600})
		}
		// Non-sensitive display name so the sidebar shows who is signed in.
		disp := user.Name
		if disp == "" {
			disp = user.Email
		}
		http.SetCookie(w, &http.Cookie{Name: "gl_user", Value: url.QueryEscape(disp), Path: "/", MaxAge: 30 * 24 * 3600})

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(ctx context.Context) *repository.User {
	user, _ := ctx.Value(userContextKey).(*repository.User)
	return user
}

// currentUserID returns the authenticated user's id (0 if absent).
func currentUserID(ctx context.Context) int {
	if u := GetUser(ctx); u != nil {
		return u.ID
	}
	return 0
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in?
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		if _, err := h.userRepo.GetUserBySession(r.Context(), cookie.Value); err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
	h.render(w, "login", map[string]any{"Error": "", "RegistrationOpen": auth.RegistrationOpen()})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if !auth.LoginLimiter.Allow(auth.ClientIP(r)) {
		h.render(w, "login", map[string]any{"Error": "Слишком много попыток входа. Подождите несколько минут.", "RegistrationOpen": auth.RegistrationOpen()})
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), email)
	if err != nil || !h.userRepo.CheckPassword(user, password) {
		h.render(w, "login", map[string]any{"Error": "Неверный email или пароль", "RegistrationOpen": auth.RegistrationOpen()})
		return
	}
	if user.Blocked {
		h.render(w, "login", map[string]any{"Error": "Аккаунт заблокирован — обратитесь к администратору", "RegistrationOpen": auth.RegistrationOpen()})
		return
	}

	token, err := h.userRepo.CreateSession(r.Context(), user.ID)
	if err != nil {
		h.render(w, "login", map[string]any{"Error": "Ошибка сервера", "RegistrationOpen": auth.RegistrationOpen()})
		return
	}

	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if !auth.RegistrationOpen() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.render(w, "register", map[string]string{"Error": ""})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if !auth.RegistrationOpen() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if name == "" || email == "" || len(password) < 6 {
		h.render(w, "register", map[string]string{"Error": "Заполните все поля (пароль мин. 6 символов)"})
		return
	}

	user, err := h.userRepo.Create(r.Context(), email, password, name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			h.render(w, "register", map[string]string{"Error": "Email уже зарегистрирован"})
			return
		}
		h.render(w, "register", map[string]string{"Error": "Ошибка сервера"})
		return
	}

	token, _ := h.userRepo.CreateSession(r.Context(), user.ID)
	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.userRepo.DeleteSession(r.Context(), cookie.Value)
	}
	auth.ClearAuthCookies(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
