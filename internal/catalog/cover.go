package catalog

import (
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/backendraz/golearn/internal/model"
)

// Gradient returns the two-stop cover gradient for a category or accent key.
func Gradient(key string) (from, to string) {
	switch key {
	case "Linux":
		return "#f7b733", "#fc4a1a"
	case "Docker":
		return "#2496ed", "#1d63ed"
	case "Kubernetes":
		return "#326ce5", "#7aa2f7"
	case "Git":
		return "#f05133", "#f0651f"
	case "Backend":
		return "#3fb950", "#2ea043"
	case "Security":
		return "#f85149", "#da3633"
	case "Основы":
		return "#a78bfa", "#7c3aed"
	case "Database":
		return "#36c5f0", "#1f6feb"
	case "Golang":
		return "#00add8", "#5dc9e2"
	default:
		return "#6ea8ff", "#b98cff"
	}
}

// DecodeDataURI parses "data:<mime>;base64,<data>".
func DecodeDataURI(uri string) (mime string, data []byte, ok bool) {
	const b64 = ";base64,"
	i := strings.Index(uri, b64)
	if i < 0 || !strings.HasPrefix(uri, "data:") {
		return "", nil, false
	}
	raw, err := base64.StdEncoding.DecodeString(uri[i+len(b64):])
	if err != nil {
		return "", nil, false
	}
	return uri[len("data:"):i], raw, true
}

const svgGrid = `<pattern id="grid" width="34" height="34" patternUnits="userSpaceOnUse"><path d="M34 0H0V34" fill="none" stroke="#ffffff" stroke-opacity="0.08" stroke-width="1"/></pattern>`

func svgOpen(b *strings.Builder, from, to string) {
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="340" viewBox="0 0 600 340" role="img"><defs>`)
	fmt.Fprintf(b, `<linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, from, to)
}

// CourseCoverSVG renders the generated course banner: gradient, grid, glow, icon and category chip.
func CourseCoverSVG(m model.Module) string {
	cat := Category(m)
	accent := m.Accent
	if accent == "" {
		accent = cat
	}
	from, to := Gradient(accent)
	var b strings.Builder
	svgOpen(&b, from, to)
	b.WriteString(`<radialGradient id="glow" cx="78%" cy="22%" r="65%"><stop offset="0" stop-color="#ffffff" stop-opacity="0.30"/><stop offset="1" stop-color="#ffffff" stop-opacity="0"/></radialGradient>`)
	b.WriteString(svgGrid + `</defs>`)
	b.WriteString(`<rect width="600" height="340" fill="url(#g)"/><rect width="600" height="340" fill="url(#grid)"/><rect width="600" height="340" fill="url(#glow)"/>`)
	fmt.Fprintf(&b, `<text x="300" y="186" font-size="150" text-anchor="middle" dominant-baseline="middle">%s</text>`, CategoryIcon(cat))
	fmt.Fprintf(&b, `<rect x="30" y="28" width="%d" height="38" rx="19" fill="#000000" fill-opacity="0.30"/>`, 30+len([]rune(cat))*14)
	fmt.Fprintf(&b, `<text x="48" y="53" font-family="-apple-system,Segoe UI,Roboto,sans-serif" font-size="19" font-weight="700" fill="#ffffff">%s</text>`, html.EscapeString(cat))
	b.WriteString(`</svg>`)
	return b.String()
}

// SpecCoverSVG renders the generated specialization banner: gradient, grid and icon.
func SpecCoverSVG(s model.Specialization) string {
	key := map[string]string{"devops": "DevOps", "golang": "Golang", "security": "Security", "database": "Database"}[s.Slug]
	from, to := Gradient(key)
	icon := s.Icon
	if icon == "" {
		icon = "📚"
	}
	var b strings.Builder
	svgOpen(&b, from, to)
	b.WriteString(svgGrid + `</defs>`)
	b.WriteString(`<rect width="600" height="340" fill="url(#g)"/><rect width="600" height="340" fill="url(#grid)"/>`)
	fmt.Fprintf(&b, `<text x="300" y="186" font-size="150" text-anchor="middle" dominant-baseline="middle">%s</text></svg>`, html.EscapeString(icon))
	return b.String()
}

// ServeCover writes an uploaded data-URI image, redirects to an external URL, or falls back to the generated SVG.
func ServeCover(w http.ResponseWriter, r *http.Request, image, fallbackSVG string) {
	if image != "" {
		if !strings.HasPrefix(image, "data:") {
			http.Redirect(w, r, image, http.StatusFound)
			return
		}
		if mime, raw, ok := DecodeDataURI(image); ok && strings.HasPrefix(mime, "image/") {
			w.Header().Set("Content-Type", mime)
			w.Header().Set("Cache-Control", "public, max-age=86400")
			_, _ = w.Write(raw)
			return
		}
	}
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(fallbackSVG))
}
