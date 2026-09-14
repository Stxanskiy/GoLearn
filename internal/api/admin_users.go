package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/repository"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const defaultUsersPage, maxUsersPage = 50, 200

func (a *API) adminListUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, before := defaultUsersPage, 0
	fields := map[string]string{}
	if v := query.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxUsersPage {
			fields["limit"] = fieldInvalidValue
		}
		limit = n
	}
	if v := query.Get("cursor"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fields["cursor"] = fieldInvalidFormat
		}
		before = n
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	ctx := r.Context()
	users, err := a.Users.ListPage(ctx, strings.TrimSpace(query.Get("q")), before, limit+1)
	if err != nil {
		a.internalError(w, "admin: list users", err)
		return
	}
	page := apigen.AdminUserPage{Items: make([]apigen.AdminUser, 0, len(users))}
	if len(users) > limit {
		users = users[:limit]
		next := strconv.Itoa(users[limit-1].ID)
		page.NextCursor = &next
	}
	for _, u := range users {
		page.Items = append(page.Items, toAdminUser(u, userFrom(ctx).ID))
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *API) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string      `json:"name"`
		Email    string      `json:"email"`
		Password string      `json:"password"`
		Role     apigen.Role `json:"role"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	name, email := strings.TrimSpace(body.Name), strings.TrimSpace(body.Email)
	fields := validateRegistration(name, email, body.Password)
	if !body.Role.Valid() {
		fields["role"] = fieldInvalidValue
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	ctx := r.Context()
	u, err := a.Users.CreateWithRole(ctx, email, body.Password, name, string(body.Role))
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeEmailTaken, "email already registered")
		return
	}
	if err != nil {
		a.internalError(w, "admin: create user", err)
		return
	}
	writeJSON(w, http.StatusCreated, toAdminUser(*u, userFrom(ctx).ID))
}

func (a *API) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	target, ok := a.adminTargetUser(w, r)
	if !ok {
		return
	}
	var body apigen.AdminUpdateUserJSONBody
	if !decodeJSON(w, r, &body) {
		return
	}
	switch {
	case body.Role == nil && body.Blocked == nil:
		writeValidation(w, map[string]string{"role": fieldRequired})
		return
	case body.Role != nil && !body.Role.Valid():
		writeValidation(w, map[string]string{"role": fieldInvalidValue})
		return
	}
	role, blocked := target.Role, target.Blocked
	if body.Role != nil {
		role = string(*body.Role)
	}
	if body.Blocked != nil {
		blocked = *body.Blocked
	}
	ctx := r.Context()
	if !notSelf(w, r, target) {
		return
	}
	if role != repository.RoleAdmin && auth.IsEnvAdmin(target.Email) {
		writeError(w, http.StatusConflict, codeRolePinnedByEnv, "email is listed in ADMIN_EMAILS")
		return
	}
	if err := a.Users.SetAccess(ctx, target.ID, role, blocked); err != nil {
		a.adminChangeError(w, "admin: set access", err)
		return
	}
	updated, err := a.Users.GetByID(ctx, target.ID)
	if err != nil {
		a.internalError(w, "admin: reload user", err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminUser(*updated, userFrom(ctx).ID))
}

func (a *API) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	target, ok := a.adminTargetUser(w, r)
	if !ok || !notSelf(w, r, target) {
		return
	}
	if err := a.Users.DeleteGuarded(r.Context(), target.ID); err != nil {
		a.adminChangeError(w, "admin: delete user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminSetUserPassword(w http.ResponseWriter, r *http.Request) {
	target, ok := a.adminTargetUser(w, r)
	if !ok {
		return
	}
	var body apigen.AdminSetUserPasswordJSONBody
	if !decodeJSON(w, r, &body) {
		return
	}
	switch {
	case utf8.RuneCountInString(body.Password) < minPasswordLen:
		writeValidation(w, map[string]string{"password": fieldTooShort})
		return
	case len(body.Password) > maxPasswordLen:
		writeValidation(w, map[string]string{"password": fieldTooLong})
		return
	}
	ctx := r.Context()
	if err := a.Users.SetPassword(ctx, target.ID, body.Password); err != nil {
		a.internalError(w, "admin: set password", err)
		return
	}
	if err := a.Users.DeleteSessions(ctx, target.ID); err != nil {
		a.internalError(w, "admin: revoke sessions", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// adminTargetUser loads the userId user, writing 404/500 itself when it returns false.
func (a *API) adminTargetUser(w http.ResponseWriter, r *http.Request) (*repository.User, bool) {
	id, ok := pathID(w, r, "userId", "user not found")
	if !ok {
		return nil, false
	}
	u, err := a.Users.GetByID(r.Context(), id)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "user not found")
		return nil, false
	}
	if err != nil {
		a.internalError(w, "admin: load user", err)
		return nil, false
	}
	return u, true
}

// notSelf rejects account changes to the signed-in admin.
func notSelf(w http.ResponseWriter, r *http.Request, target *repository.User) bool {
	if target.ID == userFrom(r.Context()).ID {
		writeError(w, http.StatusConflict, codeSelfModification, "cannot change your own account here")
		return false
	}
	return true
}

func (a *API) adminChangeError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, repository.ErrLastAdmin):
		writeError(w, http.StatusConflict, codeLastAdmin, "the platform needs an active admin")
	case isNotFound(err):
		writeError(w, http.StatusNotFound, codeNotFound, "user not found")
	default:
		a.internalError(w, op, err)
	}
}

func toAdminUser(u repository.User, selfID int) apigen.AdminUser {
	role := apigen.Role(u.Role)
	if !role.Valid() {
		role = apigen.RoleStudent
	}
	return apigen.AdminUser{
		ID: u.ID, Name: u.Name, Email: openapi_types.Email(u.Email), Role: role, Blocked: u.Blocked,
		CreatedAt: u.CreatedAt, IsSelf: u.ID == selfID, RolePinnedByEnv: auth.IsEnvAdmin(u.Email),
	}
}
