package handler

import (
	"html/template"

	"github.com/backendraz/golearn/internal/content"
)

// RenderContent turns stored content into safe display HTML for templates.
func RenderContent(format, raw string) template.HTML {
	return template.HTML(content.Render(format, raw))
}
