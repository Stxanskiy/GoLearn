package api

import (
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/go-chi/chi/v5"
)

func (a *API) getLanding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := a.Modules.Stats(ctx)
	if err != nil {
		a.internalError(w, "landing: stats", err)
		return
	}
	modules, err := a.Modules.GetAll(ctx)
	if err != nil {
		a.internalError(w, "landing: modules", err)
		return
	}
	specs, err := a.Specs.ListPublished(ctx)
	if err != nil {
		a.internalError(w, "landing: specs", err)
		return
	}
	perSpec := make(map[string]int)
	for _, m := range modules {
		perSpec[catalog.SpecForTrack(m.Track)]++
	}
	tracks := []apigen.LandingTrack{}
	for _, s := range specs {
		if n := perSpec[s.Slug]; n > 0 && s.Slug != catalog.GymSpec {
			tracks = append(tracks, apigen.LandingTrack{Slug: s.Slug, Name: s.Name, Icon: s.Icon, Description: s.Description, CoursesCount: n})
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, apigen.Landing{
		Stats: apigen.LandingStats{
			Courses:          stats.Courses,
			Lessons:          stats.Lessons,
			Labs:             stats.Labs,
			AutoCheckedTasks: stats.AutoChecked,
		},
		Tracks: tracks,
	})
}

func (a *API) getCourseCover(w http.ResponseWriter, r *http.Request) {
	m, err := a.Modules.GetBySlug(r.Context(), chi.URLParam(r, "courseSlug"))
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "course not found")
		return
	}
	if err != nil {
		a.internalError(w, "course cover", err)
		return
	}
	catalog.ServeCover(w, r, m.CoverImage, catalog.CourseCoverSVG(*m))
}

func (a *API) getSpecializationCover(w http.ResponseWriter, r *http.Request) {
	s, err := a.Specs.Get(r.Context(), chi.URLParam(r, "specSlug"))
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "specialization not found")
		return
	}
	if err != nil {
		a.internalError(w, "specialization cover", err)
		return
	}
	catalog.ServeCover(w, r, s.CoverImage, catalog.SpecCoverSVG(*s))
}
