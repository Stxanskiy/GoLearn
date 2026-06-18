package handler

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	ghtml "github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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

// RenderContent turns stored content into safe display HTML. format "md" runs the
// Markdown pipeline; anything else is treated as trusted HTML (seed/imported).
func RenderContent(format, raw string) template.HTML {
	if format == "md" {
		return renderMarkdown(raw)
	}
	return template.HTML(raw)
}
