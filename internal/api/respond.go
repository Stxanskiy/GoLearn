package api

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
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
)

// Field-level validation codes.
const (
	fieldRequired      = "required"
	fieldTooShort      = "too_short"
	fieldTooLong       = "too_long"
	fieldInvalidFormat = "invalid_format"
)

const maxJSONBody = 1 << 20

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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

// decodeJSON reads a JSON body into dst; on failure it writes the error response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || ct != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMedia, "expected application/json body")
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody)).Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, "body exceeds 1 MiB")
			return false
		}
		writeError(w, http.StatusBadRequest, codeMalformedJSON, err.Error())
		return false
	}
	return true
}
