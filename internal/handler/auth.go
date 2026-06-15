package handler

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/backendraz/golearn/internal/repository"
)

type contextKey string

const userContextKey contextKey = "user"

// registrationOpen reports whether self-registration is enabled.
// Disabled by default; set REGISTRATION_OPEN=true to allow sign-ups.
func registrationOpen() bool {
	return os.Getenv("REGISTRATION_OPEN") == "true"
}

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

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(ctx context.Context) *repository.User {
	user, _ := ctx.Value(userContextKey).(*repository.User)
	return user
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
	h.render(w, "login", map[string]any{"Error": "", "RegistrationOpen": registrationOpen()})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	user, err := h.userRepo.GetByEmail(r.Context(), email)
	if err != nil || !h.userRepo.CheckPassword(user, password) {
		h.render(w, "login", map[string]any{"Error": "Неверный email или пароль", "RegistrationOpen": registrationOpen()})
		return
	}

	token, err := h.userRepo.CreateSession(r.Context(), user.ID)
	if err != nil {
		h.render(w, "login", map[string]any{"Error": "Ошибка сервера", "RegistrationOpen": registrationOpen()})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 3600, // 30 days
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if !registrationOpen() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.render(w, "register", map[string]string{"Error": ""})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if !registrationOpen() {
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
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 3600,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.userRepo.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
