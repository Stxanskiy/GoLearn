package handler

import (
	"html/template"
	"path/filepath"
	"testing"
)

// TestTemplatesParse ensures every page template parses with base + the same
// FuncMap the renderer uses (catches syntax errors / unknown functions offline).
func TestTemplatesParse(t *testing.T) {
	funcs := templateFuncs()

	pages, err := filepath.Glob("../templates/pages/*.html")
	if err != nil || len(pages) == 0 {
		t.Fatalf("no page templates found: %v", err)
	}
	for _, p := range pages {
		p := p
		t.Run(filepath.Base(p), func(t *testing.T) {
			if _, err := template.New("").Funcs(funcs).ParseFiles("../templates/layouts/base.html", p); err != nil {
				t.Errorf("parse %s: %v", filepath.Base(p), err)
			}
		})
	}
}
