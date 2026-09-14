package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/jackc/pgx/v5"
)

type ctxKey struct{}

// userFrom returns the signed-in user or nil.
func userFrom(ctx context.Context) *repository.User {
	u, _ := ctx.Value(ctxKey{}).(*repository.User)
	return u
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// sessionUser resolves the session cookie into the request context; anonymous requests pass through.
func (a *API) sessionUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(auth.SessionCookie)
		if err != nil || c.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		user, err := a.Users.GetUserBySession(r.Context(), c.Value)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			auth.ClearAuthCookies(w)
		case err != nil:
			a.log.Error("resolve session", "error", err)
		default:
			auth.PromoteEnvAdmin(r.Context(), a.Users, user)
			r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, user))
		}
		next.ServeHTTP(w, r)
	})
}

func requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userFrom(r.Context()) == nil {
			writeError(w, http.StatusUnauthorized, codeUnauthorized, "sign in required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// csrfGuard rejects cross-site mutating requests by Origin, falling back to Sec-Fetch-Site.
func (a *API) csrfGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			if !a.originAllowed(r, origin) {
				writeError(w, http.StatusForbidden, codeCSRFRejected, "cross-origin request rejected")
				return
			}
		} else if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			writeError(w, http.StatusForbidden, codeCSRFRejected, "cross-site request rejected")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) originAllowed(r *http.Request, origin string) bool {
	for _, o := range a.cfg.AllowedOrigins {
		if strings.EqualFold(strings.TrimRight(o, "/"), origin) {
			return true
		}
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	fwd := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0])
	return fwd != "" && strings.EqualFold(u.Host, fwd)
}
