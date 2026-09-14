// Package content renders stored lesson and task content into safe display HTML.
package content

import (
	"bytes"
	"html"
	"regexp"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	ghtml "github.com/yuin/goldmark/renderer/html"
)

// mdConv renders GFM Markdown with dark code highlighting and raw-HTML passthrough.
var mdConv = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(highlighting.WithStyle("github-dark")),
	),
	goldmark.WithRendererOptions(ghtml.WithUnsafe()),
)

// Neutralizers turn executable markup into inert text so attack examples in lessons display but never run.
var (
	activeTagRe = regexp.MustCompile(`(?i)<(/?)(script|iframe|object|embed|meta|link|base|svg|form)\b`)
	onHandlerRe = regexp.MustCompile(`(?i)\son[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	jsURIRe     = regexp.MustCompile(`(?i)javascript:`)
)

// Sanitize neutralizes active markup in trusted HTML (scripts, frames, event handlers, javascript: URLs).
func Sanitize(h string) string {
	h = activeTagRe.ReplaceAllString(h, "&lt;$1$2")
	h = onHandlerRe.ReplaceAllString(h, "")
	h = jsURIRe.ReplaceAllString(h, "javascript&#58;")
	return h
}

// Render converts content of the given format ("md" or HTML) into sanitized HTML.
func Render(format, raw string) string {
	if format != "md" {
		return Sanitize(raw)
	}
	var b bytes.Buffer
	if err := mdConv.Convert([]byte(raw), &b); err != nil {
		return "<pre>" + html.EscapeString(raw) + "</pre>"
	}
	return Sanitize(b.String())
}
