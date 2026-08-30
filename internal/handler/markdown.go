package handler

import (
	"bytes"
	"html/template"
	"regexp"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	ghtml "github.com/yuin/goldmark/renderer/html"
)

// mdConv renders Markdown with GitHub-flavoured extensions, dark code highlighting
// and raw-HTML passthrough (so admins can use <span style="color:..."> for colour).
var mdConv = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(highlighting.WithStyle("github-dark")),
	),
	goldmark.WithRendererOptions(ghtml.WithUnsafe()),
)

func renderMarkdown(s string) template.HTML {
	var b bytes.Buffer
	if err := mdConv.Convert([]byte(s), &b); err != nil {
		return template.HTML("<pre>" + template.HTMLEscapeString(s) + "</pre>")
	}
	return template.HTML(b.String())
}

// Active-markup neutralisers. Lesson content is rendered as trusted HTML, but a
// security lesson legitimately contains attack payloads as EXAMPLES (an XSS
// lesson shows `<script>document.location='http://evil.com'…`). Without this,
// that example would actually execute on page load and redirect the reader.
// We turn the executable constructs into inert text so the payload DISPLAYS as
// written but never runs.
var activeTagRe = regexp.MustCompile(`(?i)<(/?)(script|iframe|object|embed|meta|link|base|svg|form)\b`)
var onHandlerRe = regexp.MustCompile(`(?i)\son[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
var jsURIRe = regexp.MustCompile(`(?i)javascript:`)

func neutralizeActiveHTML(h string) string {
	h = activeTagRe.ReplaceAllString(h, "&lt;$1$2")
	h = onHandlerRe.ReplaceAllString(h, "")
	h = jsURIRe.ReplaceAllString(h, "javascript&#58;")
	return h
}

// RenderContent turns stored content into safe display HTML. format "md" runs the
// Markdown pipeline; anything else is treated as trusted HTML (seed/imported).
// Both paths are passed through neutralizeActiveHTML so no lesson can execute a
// script or redirect the reader.
func RenderContent(format, raw string) template.HTML {
	if format == "md" {
		return template.HTML(neutralizeActiveHTML(string(renderMarkdown(raw))))
	}
	return template.HTML(neutralizeActiveHTML(raw))
}
