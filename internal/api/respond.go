package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Error codes shared with the frontend (see Error in api/openapi.yaml).
const (
	codeUnauthorized       = "unauthorized"
	codeForbidden          = "forbidden"
	codeCSRFRejected       = "csrf_rejected"
	codeNotFound           = "not_found"
	codeMethodNotAllowed   = "method_not_allowed"
	codeMalformedJSON      = "malformed_json"
	codeValidationFailed   = "validation_failed"
	codeUnsupportedMedia   = "unsupported_media_type"
	codePayloadTooLarge    = "payload_too_large"
	codeRateLimited        = "rate_limited"
	codeInternal           = "internal_error"
	codeInvalidCredentials = "invalid_credentials"
	codeAccountBlocked     = "account_blocked"
	codeRegistrationClosed = "registration_closed"
	codeEmailTaken         = "email_taken"
	codeCompletionDerived  = "completion_derived"
	codeSandboxDisabled    = "sandbox_disabled"
	codeSandboxNotRunning  = "sandbox_not_running"
	codeSandboxError       = "sandbox_error"
	codePathOutsideJail    = "path_outside_jail"
	codeFileTooLarge       = "file_too_large"
	codeTaskNotAutoChecked = "task_not_auto_checked"
	codeTaskAutoChecked    = "task_auto_checked"
	codeTaskNotCode        = "task_not_code"
	codeSlugTaken          = "slug_taken"
	codeAlreadyAuthor      = "already_author"
	codeImportBlocked      = "import_blocked"
	codeSpecNotEmpty       = "specialization_not_empty"
	codeReviewRequired     = "review_required"
	codeReviewPending      = "review_pending"
	codeReviewNotNeeded    = "review_not_needed"
	codeReviewClosed       = "review_closed"
	codeSelfModification   = "self_modification"
	codeLastAdmin          = "last_admin"
	codeRolePinnedByEnv    = "role_pinned_by_env"
)

// Field-level validation codes.
const (
	fieldRequired      = "required"
	fieldTooShort      = "too_short"
	fieldTooLong       = "too_long"
	fieldInvalidFormat = "invalid_format"
	fieldInvalidValue  = "invalid_value"
	fieldNotFound      = "not_found"
	fieldNotAuthor     = "not_author"
)

const maxJSONBody = 1 << 20

// writeJSON encodes v before writing headers so an encoding failure becomes a 500 instead of an empty body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		slog.Error("encode response", "error", err)
		status = http.StatusInternalServerError
		body, _ = json.Marshal(apigen.Error{Error: apigen.ErrorBody{Code: codeInternal, Message: "internal error"}})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apigen.Error{Error: apigen.ErrorBody{Code: code, Message: message}})
}

// writeValidation responds 422 with field name → validation code.
func writeValidation(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, apigen.Error{Error: apigen.ErrorBody{
		Code:    codeValidationFailed,
		Message: "request validation failed",
		Details: &apigen.ErrorDetails{Fields: &fields},
	}})
}

// decodeOptionalJSON is decodeJSON that accepts an empty body.
func decodeOptionalJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.ContentLength == 0 && r.Header.Get("Content-Type") == "" {
		return true
	}
	return decodeJSON(w, r, dst)
}

// decodeJSON reads a JSON body into dst; on failure it writes the error response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return decodeJSONLimit(w, r, dst, maxJSONBody)
}

// decodeJSONLimit is decodeJSON with a custom body size limit in bytes.
func decodeJSONLimit(w http.ResponseWriter, r *http.Request, dst any, limit int64) bool {
	if ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || ct != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMedia, "expected application/json body")
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit)).Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, "body exceeds the size limit")
			return false
		}
		writeError(w, http.StatusBadRequest, codeMalformedJSON, err.Error())
		return false
	}
	return true
}

// isUniqueViolation reports whether err is a unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// isNotFound reports whether a repository error means the row does not exist.
func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
