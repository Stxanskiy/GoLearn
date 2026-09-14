package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
)

func TestWriteJSONEncodingFailure(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusCreated, apigen.AuthorRef{ID: 1, Email: "not an email"})
	if w.Code != http.StatusInternalServerError || errorCode(t, w) != codeInternal {
		t.Errorf("status %d body %q", w.Code, w.Body)
	}
}
