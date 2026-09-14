// Package auth holds session cookies, rate limits and env auth rules shared by HTML handlers and the API.
package auth

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/backendraz/golearn/internal/repository"
)

// SessionCookie is the HttpOnly cookie holding the session token.
const SessionCookie = "session"

const sessionMaxAge = 30 * 24 * 3600

// RegistrationOpen reports whether self-registration is enabled (REGISTRATION_OPEN=true).
func RegistrationOpen() bool {
	return os.Getenv("REGISTRATION_OPEN") == "true"
}

// SetSessionCookie writes the session cookie; Secure when the request came over HTTPS.
func SetSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   IsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookies removes the session cookie and the legacy navbar hint cookies.
func ClearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{SessionCookie, "gl_role", "gl_user"} {
		http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1})
	}
}

// IsHTTPS reports whether the request arrived over TLS directly or via a TLS proxy.
func IsHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// IsEnvAdmin reports whether the email is listed in ADMIN_EMAILS.
func IsEnvAdmin(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false
	}
	for _, e := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		if strings.ToLower(strings.TrimSpace(e)) == email {
			return true
		}
	}
	return false
}

// RoleSetter is the part of the user repository needed to promote env admins.
type RoleSetter interface {
	SetRole(ctx context.Context, userID int, role string) error
}

// PromoteEnvAdmin upgrades a user listed in ADMIN_EMAILS to admin, in the DB and in place.
func PromoteEnvAdmin(ctx context.Context, repo RoleSetter, user *repository.User) {
	if user.IsAdmin() || !IsEnvAdmin(user.Email) {
		return
	}
	if err := repo.SetRole(ctx, user.ID, "admin"); err == nil {
		user.Role = "admin"
	}
}
