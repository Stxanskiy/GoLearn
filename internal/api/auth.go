package api

import (
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	minPasswordLen = 6
	maxPasswordLen = 72 // bcrypt limit, bytes
	maxNameLen     = 100
	maxEmailLen    = 254
)

func (a *API) getAuthConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, apigen.AuthConfig{RegistrationOpen: auth.RegistrationOpen()})
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var body apigen.LoginJSONBody
	if !decodeJSON(w, r, &body) {
		return
	}
	email := strings.TrimSpace(body.Email)
	fields := map[string]string{}
	if email == "" {
		fields["email"] = fieldRequired
	}
	if body.Password == "" {
		fields["password"] = fieldRequired
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	if !auth.LoginLimiter.Allow(auth.ClientIP(r)) {
		writeRateLimited(w, auth.LoginLimiter)
		return
	}

	user, err := a.Users.GetByEmail(r.Context(), email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.internalError(w, "load user", err)
		return
	}
	if user == nil || !a.Users.CheckPassword(user, body.Password) {
		writeError(w, http.StatusUnauthorized, codeInvalidCredentials, "invalid email or password")
		return
	}
	if user.Blocked {
		writeError(w, http.StatusForbidden, codeAccountBlocked, "account is blocked")
		return
	}
	a.startSession(w, r, user, http.StatusOK)
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	if !auth.RegistrationOpen() {
		writeError(w, http.StatusForbidden, codeRegistrationClosed, "self-registration is disabled")
		return
	}
	var body apigen.RegisterJSONBody
	if !decodeJSON(w, r, &body) {
		return
	}
	name, email := strings.TrimSpace(body.Name), strings.TrimSpace(body.Email)
	if fields := validateRegistration(name, email, body.Password); len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	if !auth.RegisterLimiter.Allow(auth.ClientIP(r)) {
		writeRateLimited(w, auth.RegisterLimiter)
		return
	}

	user, err := a.Users.Create(r.Context(), email, body.Password, name)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeError(w, http.StatusConflict, codeEmailTaken, "email already registered")
		return
	}
	if err != nil {
		a.internalError(w, "create user", err)
		return
	}
	a.startSession(w, r, user, http.StatusCreated)
}

func validateRegistration(name, email, password string) map[string]string {
	fields := map[string]string{}
	switch n := utf8.RuneCountInString(name); {
	case n == 0:
		fields["name"] = fieldRequired
	case n > maxNameLen:
		fields["name"] = fieldTooLong
	}
	switch {
	case email == "":
		fields["email"] = fieldRequired
	case len(email) > maxEmailLen:
		fields["email"] = fieldTooLong
	default:
		if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
			fields["email"] = fieldInvalidFormat
		}
	}
	switch {
	case password == "":
		fields["password"] = fieldRequired
	case utf8.RuneCountInString(password) < minPasswordLen:
		fields["password"] = fieldTooShort
	case len(password) > maxPasswordLen:
		fields["password"] = fieldTooLong
	}
	return fields
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.SessionCookie); err == nil && c.Value != "" {
		if err := a.Users.DeleteSession(r.Context(), c.Value); err != nil {
			a.log.Error("delete session", "error", err)
		}
	}
	auth.ClearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) getMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, toMe(userFrom(r.Context())))
}

// startSession creates a session, sets the cookie and responds with the user.
func (a *API) startSession(w http.ResponseWriter, r *http.Request, user *repository.User, status int) {
	token, err := a.Users.CreateSession(r.Context(), user.ID)
	if err != nil {
		a.internalError(w, "create session", err)
		return
	}
	auth.PromoteEnvAdmin(r.Context(), a.Users, user)
	auth.SetSessionCookie(w, r, token)
	writeJSON(w, status, toMe(user))
}

func toMe(u *repository.User) apigen.Me {
	role := apigen.RoleStudent
	if u.IsAdmin() {
		role = apigen.RoleAdmin
	}
	return apigen.Me{
		ID:        u.ID,
		Name:      u.Name,
		Email:     openapi_types.Email(u.Email),
		Role:      role,
		IsAdmin:   u.IsAdmin(),
		CreatedAt: u.CreatedAt,
	}
}

func writeRateLimited(w http.ResponseWriter, l *auth.Limiter) {
	w.Header().Set("Retry-After", strconv.Itoa(int(l.Window().Seconds())))
	writeError(w, http.StatusTooManyRequests, codeRateLimited, "too many attempts")
}

func (a *API) internalError(w http.ResponseWriter, op string, err error) {
	a.log.Error(op, "error", err)
	writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
}
