package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeUsers is an in-memory userStore; passwords are stored in PasswordHash as plain text.
type fakeUsers struct {
	byEmail  map[string]*repository.User
	sessions map[string]int
	nextID   int
}

func newFakeUsers(users ...*repository.User) *fakeUsers {
	f := &fakeUsers{byEmail: map[string]*repository.User{}, sessions: map[string]int{}, nextID: 100}
	for _, u := range users {
		f.byEmail[u.Email] = u
	}
	return f
}

func (f *fakeUsers) GetByEmail(_ context.Context, email string) (*repository.User, error) {
	if u, ok := f.byEmail[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeUsers) CheckPassword(u *repository.User, password string) bool {
	return u.PasswordHash == password
}

func (f *fakeUsers) Create(_ context.Context, email, password, name string) (*repository.User, error) {
	if _, ok := f.byEmail[email]; ok {
		return nil, &pgconn.PgError{Code: "23505"}
	}
	f.nextID++
	u := &repository.User{ID: f.nextID, Email: email, PasswordHash: password, Name: name, Role: "student", CreatedAt: time.Now()}
	f.byEmail[email] = u
	return u, nil
}

func (f *fakeUsers) CreateSession(_ context.Context, userID int) (string, error) {
	token := "tok-" + time.Now().Format("150405.000000000")
	f.sessions[token] = userID
	return token, nil
}

func (f *fakeUsers) GetUserBySession(_ context.Context, token string) (*repository.User, error) {
	id, ok := f.sessions[token]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	for _, u := range f.byEmail {
		if u.ID == id && !u.Blocked {
			cp := *u
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeUsers) DeleteSession(_ context.Context, token string) error {
	delete(f.sessions, token)
	return nil
}

func (f *fakeUsers) SetRole(_ context.Context, userID int, role string) error {
	for _, u := range f.byEmail {
		if u.ID == userID {
			u.Role = role
		}
	}
	return nil
}

func newTestAPI(t *testing.T, users *fakeUsers) http.Handler {
	t.Helper()
	return newTestAPIWith(t, users, newFakeContent())
}

func newTestAPIWith(t *testing.T, users *fakeUsers, content *fakeContent) http.Handler {
	t.Helper()
	auth.LoginLimiter = auth.NewLimiter(10, time.Minute)
	auth.RegisterLimiter = auth.NewLimiter(5, time.Minute)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(content.stores(users), Config{AllowedOrigins: []string{"http://localhost:3000"}}, log).Routes()
}

type reqOpt func(*http.Request)

func withCookie(token string) reqOpt {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: token}) }
}

func withHeader(k, v string) reqOpt {
	return func(r *http.Request) { r.Header.Set(k, v) }
}

func do(h http.Handler, method, path, body string, opts ...reqOpt) *httptest.ResponseRecorder {
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	r := httptest.NewRequest(method, path, rd)
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for _, o := range opts {
		o(r)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var e apigen.Error
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode error body %q: %v", w.Body.String(), err)
	}
	return e.Error.Code
}

func sessionCookie(w *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.SessionCookie && c.Value != "" {
			return c
		}
	}
	return nil
}

func alice() *repository.User {
	return &repository.User{ID: 1, Email: "alice@example.com", PasswordHash: "secret1", Name: "Alice", Role: "student"}
}

func TestLogin(t *testing.T) {
	blocked := &repository.User{ID: 2, Email: "bob@example.com", PasswordHash: "secret1", Blocked: true}
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"success", `{"email":" alice@example.com ","password":"secret1"}`, http.StatusOK, ""},
		{"wrong password", `{"email":"alice@example.com","password":"nope"}`, http.StatusUnauthorized, codeInvalidCredentials},
		{"unknown email", `{"email":"x@example.com","password":"secret1"}`, http.StatusUnauthorized, codeInvalidCredentials},
		{"blocked", `{"email":"bob@example.com","password":"secret1"}`, http.StatusForbidden, codeAccountBlocked},
		{"missing fields", `{}`, http.StatusUnprocessableEntity, codeValidationFailed},
		{"malformed json", `{"email":`, http.StatusBadRequest, codeMalformedJSON},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPI(t, newFakeUsers(alice(), blocked))
			w := do(h, http.MethodPost, "/auth/login", tt.body)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantCode != "" {
				if got := errorCode(t, w); got != tt.wantCode {
					t.Errorf("code = %q, want %q", got, tt.wantCode)
				}
				if sessionCookie(w) != nil {
					t.Error("session cookie set on failure")
				}
				return
			}
			c := sessionCookie(w)
			if c == nil || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
				t.Fatalf("bad session cookie: %+v", c)
			}
			var me apigen.Me
			if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil || me.ID != 1 || me.IsAdmin {
				t.Errorf("me = %+v, err %v", me, err)
			}
		})
	}
}

func TestLoginRequiresJSON(t *testing.T) {
	h := newTestAPI(t, newFakeUsers(alice()))
	w := do(h, http.MethodPost, "/auth/login", `email=a`, withHeader("Content-Type", "application/x-www-form-urlencoded"))
	if w.Code != http.StatusUnsupportedMediaType || errorCode(t, w) != codeUnsupportedMedia {
		t.Fatalf("status = %d body %s", w.Code, w.Body)
	}
}

func TestLoginRateLimited(t *testing.T) {
	h := newTestAPI(t, newFakeUsers(alice()))
	auth.LoginLimiter = auth.NewLimiter(2, time.Minute)
	for i := 0; i < 2; i++ {
		do(h, http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"nope"}`)
	}
	w := do(h, http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"secret1"}`)
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") != "60" {
		t.Fatalf("status = %d, Retry-After %q", w.Code, w.Header().Get("Retry-After"))
	}
}

func TestLoginPromotesEnvAdmin(t *testing.T) {
	t.Setenv("ADMIN_EMAILS", "other@example.com, ALICE@example.com")
	users := newFakeUsers(alice())
	h := newTestAPI(t, users)
	w := do(h, http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"secret1"}`)
	var me apigen.Me
	_ = json.Unmarshal(w.Body.Bytes(), &me)
	if !me.IsAdmin || me.Role != apigen.RoleAdmin || users.byEmail["alice@example.com"].Role != "admin" {
		t.Fatalf("not promoted: %+v", me)
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name       string
		open       bool
		body       string
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{"closed", false, `{"name":"N","email":"n@example.com","password":"secret1"}`, http.StatusForbidden, codeRegistrationClosed, ""},
		{"success", true, `{"name":" Neo ","email":"neo@example.com","password":"secret1"}`, http.StatusCreated, "", ""},
		{"duplicate", true, `{"name":"A","email":"alice@example.com","password":"secret1"}`, http.StatusConflict, codeEmailTaken, ""},
		{"bad email", true, `{"name":"A","email":"Alice <a@example.com>","password":"secret1"}`, http.StatusUnprocessableEntity, codeValidationFailed, "email"},
		{"short password", true, `{"name":"A","email":"a@example.com","password":"12345"}`, http.StatusUnprocessableEntity, codeValidationFailed, "password"},
		{"password over bcrypt limit", true, `{"name":"A","email":"a@example.com","password":"` + strings.Repeat("x", 73) + `"}`, http.StatusUnprocessableEntity, codeValidationFailed, "password"},
		{"empty name", true, `{"name":"  ","email":"a@example.com","password":"secret1"}`, http.StatusUnprocessableEntity, codeValidationFailed, "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.open {
				t.Setenv("REGISTRATION_OPEN", "true")
			}
			h := newTestAPI(t, newFakeUsers(alice()))
			w := do(h, http.MethodPost, "/auth/register", tt.body)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantCode == "" {
				if sessionCookie(w) == nil {
					t.Error("no session cookie after register")
				}
				return
			}
			var e apigen.Error
			_ = json.Unmarshal(w.Body.Bytes(), &e)
			if e.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", e.Error.Code, tt.wantCode)
			}
			if tt.wantField != "" && (e.Error.Details == nil || e.Error.Details.Fields == nil || (*e.Error.Details.Fields)[tt.wantField] == "") {
				t.Errorf("missing field error for %q: %s", tt.wantField, w.Body)
			}
		})
	}
}

func TestMeAndLogout(t *testing.T) {
	users := newFakeUsers(alice())
	h := newTestAPI(t, users)

	if w := do(h, http.MethodGet, "/me", ""); w.Code != http.StatusUnauthorized || errorCode(t, w) != codeUnauthorized {
		t.Fatalf("anonymous /me: %d %s", w.Code, w.Body)
	}

	c := sessionCookie(do(h, http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"secret1"}`))
	w := do(h, http.MethodGet, "/me", "", withCookie(c.Value))
	if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("/me: %d cache %q", w.Code, w.Header().Get("Cache-Control"))
	}

	if w := do(h, http.MethodPost, "/auth/logout", "", withCookie(c.Value)); w.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", w.Code)
	}
	if len(users.sessions) != 0 {
		t.Error("session not deleted")
	}
	w = do(h, http.MethodGet, "/me", "", withCookie(c.Value))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("/me after logout: %d", w.Code)
	}
	if !cookieCleared(w, auth.SessionCookie) {
		t.Error("stale session cookie not cleared")
	}
}

func cookieCleared(w *httptest.ResponseRecorder, name string) bool {
	for _, c := range w.Result().Cookies() {
		if c.Name == name && c.MaxAge < 0 {
			return true
		}
	}
	return false
}

func TestCSRFGuard(t *testing.T) {
	body := `{"email":"alice@example.com","password":"secret1"}`
	tests := []struct {
		name   string
		opts   []reqOpt
		wantOK bool
	}{
		{"no browser headers", nil, true},
		{"same host origin", []reqOpt{withHeader("Origin", "http://example.com")}, true},
		{"forwarded host origin", []reqOpt{withHeader("Origin", "https://learn.prod-factory.ru"), withHeader("X-Forwarded-Host", "learn.prod-factory.ru")}, true},
		{"allowlisted origin", []reqOpt{withHeader("Origin", "http://localhost:3000")}, true},
		{"foreign origin", []reqOpt{withHeader("Origin", "https://evil.test")}, false},
		{"null origin", []reqOpt{withHeader("Origin", "null")}, false},
		{"same-origin fetch", []reqOpt{withHeader("Sec-Fetch-Site", "same-origin")}, true},
		{"cross-site fetch", []reqOpt{withHeader("Sec-Fetch-Site", "cross-site")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPI(t, newFakeUsers(alice()))
			w := do(h, http.MethodPost, "/auth/login", body, tt.opts...)
			rejected := w.Code == http.StatusForbidden && errorCode(t, w) == codeCSRFRejected
			if rejected == tt.wantOK {
				t.Fatalf("status = %d, wantOK %v (%s)", w.Code, tt.wantOK, w.Body)
			}
		})
	}
	h := newTestAPI(t, newFakeUsers(alice()))
	if w := do(h, http.MethodGet, "/auth/config", "", withHeader("Origin", "https://evil.test")); w.Code != http.StatusOK {
		t.Errorf("GET must not be CSRF-guarded: %d", w.Code)
	}
}

func TestUnknownRouteIsJSON(t *testing.T) {
	h := newTestAPI(t, newFakeUsers())
	w := do(h, http.MethodGet, "/nope", "")
	if w.Code != http.StatusNotFound || errorCode(t, w) != codeNotFound {
		t.Fatalf("404: %d %s", w.Code, w.Body)
	}
	w = do(h, http.MethodDelete, "/auth/config", "")
	if w.Code != http.StatusMethodNotAllowed || errorCode(t, w) != codeMethodNotAllowed {
		t.Fatalf("405: %d %s", w.Code, w.Body)
	}
}
