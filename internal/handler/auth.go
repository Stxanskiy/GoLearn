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

// setSessionCookie writes the session cookie with secure defaults. Secure is
// enabled when the request arrived over HTTPS (directly or via a TLS proxy).
func setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	secure := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 3600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// adminEmails returns the set of emails granted admin via ADMIN_EMAILS env.
func adminEmails() map[string]bool {
	set := make(map[string]bool)
	for _, e := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		if e = strings.ToLower(strings.TrimSpace(e)); e != "" {
			set[e] = true
		}
	}
	return set
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

		// Auto-promote configured emails to admin (ADMIN_EMAILS=a@x,b@y).
		if !user.IsAdmin() && adminEmails()[strings.ToLower(user.Email)] {
			if err := h.userRepo.SetRole(r.Context(), user.ID, "admin"); err == nil {
				user.Role = "admin"
			}
		}

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

	setSessionCookie(w, r, token)
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
	setSessionCookie(w, r, token)
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
