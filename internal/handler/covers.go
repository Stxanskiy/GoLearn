package handler

import (
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// gradientFor returns the two-stop gradient (from, to) for a category/accent key.
func gradientFor(key string) (string, string) {
	switch key {
	case "Linux":
		return "#f7b733", "#fc4a1a"
	case "Docker":
		return "#2496ed", "#1d63ed"
	case "Kubernetes":
		return "#326ce5", "#7aa2f7"
	case "Git":
		return "#f05133", "#f0651f"
	case "DevOps":
		return "#6ea8ff", "#b98cff"
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

// deriveLabel maps a module difficulty to a course-type label (devops404 style).
func deriveLabel(difficulty string) string {
	switch difficulty {
	case "beginner":
		return "Старт"
	case "intermediate":
		return "Практика"
	case "advanced", "expert":
		return "Вызов"
	default:
		return "Старт"
	}
}

// decodeDataURI parses "data:<mime>;base64,<data>" into mime + bytes.
func decodeDataURI(uri string) (mime string, data []byte, ok bool) {
	const b64 = ";base64,"
	i := strings.Index(uri, b64)
	if i < 0 || !strings.HasPrefix(uri, "data:") {
		return "", nil, false
	}
	mime = uri[len("data:"):i]
	raw, err := base64.StdEncoding.DecodeString(uri[i+len(b64):])
	if err != nil {
		return "", nil, false
	}
	return mime, raw, true
}

// CourseCover renders a generated SVG cover for a course (used when no real photo).
func (h *Handler) CourseCover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	mod, err := h.moduleRepo.GetBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Real cover set: data URI -> decode and serve bytes; otherwise redirect.
	if mod.CoverImage != "" {
		if strings.HasPrefix(mod.CoverImage, "data:") {
			if mime, raw, ok := decodeDataURI(mod.CoverImage); ok {
				w.Header().Set("Content-Type", mime)
				w.Header().Set("Cache-Control", "public, max-age=86400")
				_, _ = w.Write(raw)
				return
			}
		} else {
			http.Redirect(w, r, mod.CoverImage, http.StatusFound)
			return
		}
	}

	cat := mod.Category
	if cat == "" {
		cat = categorize(mod.Track, mod.Title, mod.Slug)
	}
	accent := mod.Accent
	if accent == "" {
		accent = cat
	}
	from, to := gradientFor(accent)
	icon := categoryIcon(cat)

	// Clean banner: gradient + texture + big centered icon + category chip.
	// No title text — the card body already shows the title (avoids overlap).
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="340" viewBox="0 0 600 340" role="img">`)
	b.WriteString(`<defs>`)
	fmt.Fprintf(&b, `<linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, from, to)
	b.WriteString(`<radialGradient id="glow" cx="78%" cy="22%" r="65%"><stop offset="0" stop-color="#ffffff" stop-opacity="0.30"/><stop offset="1" stop-color="#ffffff" stop-opacity="0"/></radialGradient>`)
	b.WriteString(`<pattern id="grid" width="34" height="34" patternUnits="userSpaceOnUse"><path d="M34 0H0V34" fill="none" stroke="#ffffff" stroke-opacity="0.08" stroke-width="1"/></pattern>`)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="600" height="340" fill="url(#g)"/>`)
	b.WriteString(`<rect width="600" height="340" fill="url(#grid)"/>`)
	b.WriteString(`<rect width="600" height="340" fill="url(#glow)"/>`)
	// Big centered icon
	fmt.Fprintf(&b, `<text x="300" y="186" font-size="150" text-anchor="middle" dominant-baseline="middle">%s</text>`, icon)
	// Category chip, top-left
	chipW := 30 + len([]rune(cat))*14
	fmt.Fprintf(&b, `<rect x="30" y="28" width="%d" height="38" rx="19" fill="#000000" fill-opacity="0.30"/>`, chipW)
	fmt.Fprintf(&b, `<text x="%d" y="53" font-family="-apple-system,Segoe UI,Roboto,sans-serif" font-size="19" font-weight="700" fill="#ffffff">%s</text>`, 48, html.EscapeString(cat))
	b.WriteString(`</svg>`)

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(b.String()))
}
