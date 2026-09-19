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

// Generated covers keep the 16:9 contract of the cover slot.
const (
	coverWidth  = 1280
	coverHeight = 720
)

const svgGrid = `<pattern id="grid" width="72" height="72" patternUnits="userSpaceOnUse"><path d="M72 0H0V72" fill="none" stroke="#ffffff" stroke-opacity="0.08" stroke-width="2"/></pattern>`

func svgOpen(b *strings.Builder, from, to string) {
	fmt.Fprintf(b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img"><defs>`,
		coverWidth, coverHeight, coverWidth, coverHeight)
	fmt.Fprintf(b, `<linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, from, to)
}

// svgBackdrop paints the gradient, the grid and any extra overlay fills.
func svgBackdrop(b *strings.Builder, overlays ...string) {
	fmt.Fprintf(b, `<rect width="%d" height="%d" fill="url(#g)"/>`, coverWidth, coverHeight)
	fmt.Fprintf(b, `<rect width="%d" height="%d" fill="url(#grid)"/>`, coverWidth, coverHeight)
	for _, id := range overlays {
		fmt.Fprintf(b, `<rect width="%d" height="%d" fill="url(#%s)"/>`, coverWidth, coverHeight, id)
	}
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
	svgBackdrop(&b, "glow")
	fmt.Fprintf(&b, `<text x="640" y="394" font-size="320" text-anchor="middle" dominant-baseline="middle">%s</text>`, CategoryIcon(cat))
	fmt.Fprintf(&b, `<rect x="64" y="60" width="%d" height="80" rx="40" fill="#000000" fill-opacity="0.30"/>`, 64+len([]rune(cat))*30)
	fmt.Fprintf(&b, `<text x="102" y="113" font-family="-apple-system,Segoe UI,Roboto,sans-serif" font-size="40" font-weight="700" fill="#ffffff">%s</text>`, html.EscapeString(cat))
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
	svgBackdrop(&b)
	fmt.Fprintf(&b, `<text x="640" y="394" font-size="320" text-anchor="middle" dominant-baseline="middle">%s</text></svg>`, html.EscapeString(icon))
	return b.String()
}

// ServeCover writes an uploaded data-URI image, redirects to an external URL, or falls back to the generated SVG.
func ServeCover(w http.ResponseWriter, r *http.Request, image, fallbackSVG string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'; sandbox")
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
