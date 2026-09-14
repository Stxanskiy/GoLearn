package handler

import (
	"net/http"

	"github.com/backendraz/golearn/internal/catalog"
	"github.com/go-chi/chi/v5"
)

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

// SpecCover serves a specialization's cover (real image) or a generated SVG.
func (h *Handler) SpecCover(w http.ResponseWriter, r *http.Request) {
	s, err := h.specRepo.Get(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	catalog.ServeCover(w, r, s.CoverImage, catalog.SpecCoverSVG(*s))
}

// CourseCover serves a course's cover (real image) or a generated SVG.
func (h *Handler) CourseCover(w http.ResponseWriter, r *http.Request) {
	mod, err := h.moduleRepo.GetBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	catalog.ServeCover(w, r, mod.CoverImage, catalog.CourseCoverSVG(*mod))
}
