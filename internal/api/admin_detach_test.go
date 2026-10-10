package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
)

// Detaching converts every Markdown lesson and hands ownership to the database.
// Both halves matter: converting without stamping edited_at would be undone by
// the next deploy, which is the whole reason they are one operation.
func TestDetachFromSeedConvertsAndStamps(t *testing.T) {
	users, c := storefront()
	h := newTestAPIWith(t, users, c)

	var mdID int
	for i := range c.lessons {
		if c.lessons[i].ID == 100 {
			c.lessons[i].Format, c.lessons[i].Content = "md", "## Заголовок\n\nабзац"
			mdID = c.lessons[i].ID
		}
	}
	if mdID == 0 {
		t.Fatal("fixture lacks the lesson this test converts")
	}

	// An author must not be able to decide this for the whole platform.
	if w := do(h, http.MethodPost, "/admin/content/detach-from-seed", "", withCookie(studentToken)); w.Code == http.StatusOK {
		t.Fatal("a student detached the whole catalogue")
	}

	w := do(h, http.MethodPost, "/admin/content/detach-from-seed", "", withCookie(adminToken))
	if w.Code != http.StatusOK {
		t.Fatalf("detach: %d %s", w.Code, w.Body)
	}
	got := decode[apigen.DetachResult](t, w)
	if got.Converted != 1 || got.LessonsDetached == 0 || got.ModulesDetached == 0 {
		t.Fatalf("result = %+v", got)
	}

	var converted string
	var format string
	for _, l := range c.lessons {
		if l.ID == mdID {
			converted, format = l.Content, l.Format
		}
	}
	if format != "html" {
		t.Errorf("format = %q, want html", format)
	}
	if !strings.Contains(converted, "<h2>") {
		t.Errorf("content was not rendered: %q", converted)
	}

	// Idempotent: nothing is left in Markdown, so a second run converts nothing.
	w = do(h, http.MethodPost, "/admin/content/detach-from-seed", "", withCookie(adminToken))
	if w.Code != http.StatusOK {
		t.Fatalf("second detach: %d %s", w.Code, w.Body)
	}
	if again := decode[apigen.DetachResult](t, w); again.Converted != 0 {
		t.Errorf("second run converted %d lessons", again.Converted)
	}
}
