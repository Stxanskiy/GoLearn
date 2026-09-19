package api

import (
	"cmp"
	"context"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

const maxLandingTracks = 50

func (a *API) getLanding(w http.ResponseWriter, r *http.Request) {
	sort, dir, limit, ok := landingQuery(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	stats, err := a.Modules.Stats(ctx)
	if err != nil {
		a.internalError(w, "landing: stats", err)
		return
	}
	tracks, err := a.landingTracks(ctx)
	if err != nil {
		a.internalError(w, "landing: tracks", err)
		return
	}
	sortTracks(tracks, sort, dir)
	if limit > 0 && limit < len(tracks) {
		tracks = tracks[:limit]
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

// landingQuery reads the sort and limit parameters, answering 422 itself when they are wrong.
func landingQuery(w http.ResponseWriter, r *http.Request) (apigen.TrackSort, apigen.SortDirection, int, bool) {
	query := r.URL.Query()
	sort, dir, limit := apigen.Order, apigen.Asc, 0
	fields := map[string]string{}

	if v := query.Get("sort"); v != "" {
		if sort = apigen.TrackSort(v); !sort.Valid() {
			fields["sort"] = fieldInvalidValue
		}
	}
	if v := query.Get("dir"); v != "" {
		if dir = apigen.SortDirection(v); !dir.Valid() {
			fields["dir"] = fieldInvalidValue
		}
	}
	if v := query.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxLandingTracks {
			fields["limit"] = fieldInvalidValue
		}
		limit = n
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return sort, dir, limit, false
	}
	return sort, dir, limit, true
}

// landingTracks lists the published specializations that have courses, with their course and lesson counts.
func (a *API) landingTracks(ctx context.Context) ([]apigen.LandingTrack, error) {
	modules, err := a.Modules.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	lessons, err := a.Lessons.ListPublishedOutline(ctx)
	if err != nil {
		return nil, err
	}
	specs, err := a.Specs.ListPublished(ctx)
	if err != nil {
		return nil, err
	}

	specOfModule := make(map[int]string, len(modules))
	courses := make(map[string]int)
	for _, m := range modules {
		if m.IsTrainer {
			continue
		}
		spec := catalog.SpecForTrack(m.Track)
		specOfModule[m.ID] = spec
		courses[spec]++
	}
	lessonCount := make(map[string]int)
	for _, l := range lessons {
		if spec, ok := specOfModule[l.ModuleID]; ok {
			lessonCount[spec]++
		}
	}

	tracks := []apigen.LandingTrack{}
	for _, s := range specs {
		if courses[s.Slug] == 0 {
			continue
		}
		tracks = append(tracks, apigen.LandingTrack{
			Slug:         s.Slug,
			Name:         s.Name,
			Icon:         s.Icon,
			IconURL:      s.IconURL,
			Description:  s.Description,
			CoursesCount: courses[s.Slug],
			LessonsCount: lessonCount[s.Slug],
		})
	}
	return tracks, nil
}

// sortTracks orders the tracks in place; `order` keeps the order set in the admin.
func sortTracks(tracks []apigen.LandingTrack, by apigen.TrackSort, dir apigen.SortDirection) {
	if by == apigen.Order {
		if dir == apigen.Desc {
			slices.Reverse(tracks)
		}
		return
	}
	compare := func(x, y apigen.LandingTrack) int {
		if by == apigen.Courses {
			return cmp.Compare(x.CoursesCount, y.CoursesCount)
		}
		return cmp.Compare(x.LessonsCount, y.LessonsCount)
	}
	slices.SortStableFunc(tracks, func(x, y apigen.LandingTrack) int {
		if dir == apigen.Desc {
			return compare(y, x)
		}
		return compare(x, y)
	})
}

// getPublicCatalog serves the catalog for anonymous visitors: published content, no progress.
func (a *API) getPublicCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	courses, specs, err := a.publicCatalog(ctx)
	if err != nil {
		a.internalError(w, "public catalog", err)
		return
	}
	filter, ok := catalogFilterFrom(w, r)
	if !ok {
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, buildCatalog(specs, courses, filter))
}

// getPublicSpecialization serves one specialization with its courses, without progress.
func (a *API) getPublicSpecialization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	spec, err := a.Specs.Get(ctx, chi.URLParam(r, "specSlug"))
	if err == nil && !spec.Published {
		err = pgx.ErrNoRows
	}
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "specialization not found")
		return
	}
	if err != nil {
		a.internalError(w, "public specialization", err)
		return
	}
	courses, _, err := a.publicCatalog(ctx)
	if err != nil {
		a.internalError(w, "public specialization: courses", err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, specWithCourses(*spec, courses, courses))
}

// publicCatalog builds every published course without progress, plus the published specializations.
func (a *API) publicCatalog(ctx context.Context) ([]catalog.Course, []model.Specialization, error) {
	courses, err := a.publishedCourses(ctx, 0)
	if err != nil {
		return nil, nil, err
	}
	specs, err := a.Specs.ListPublished(ctx)
	if err != nil {
		return nil, nil, err
	}
	return courses, specs, nil
}

// getCoursePreview serves the public course page: specialization and lessons, no progress.
func (a *API) getCoursePreview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := a.Modules.GetBySlug(ctx, chi.URLParam(r, "courseSlug"))
	if err == nil && !m.Published {
		err = pgx.ErrNoRows
	}
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "course not found")
		return
	}
	if err != nil {
		a.internalError(w, "course preview: get", err)
		return
	}
	lessons, err := a.Lessons.GetByModule(ctx, m.ID)
	if err != nil {
		a.internalError(w, "course preview: lessons", err)
		return
	}
	c := catalog.BuildCourse(*m, lessons, catalog.UserProgress{})

	links := make([]apigen.LessonLink, 0, len(lessons))
	labs := 0
	for _, l := range lessons {
		if l.Kind == "lab" {
			labs++
		}
		links = append(links, apigen.LessonLink{Slug: l.Slug, Title: l.Title, Kind: lessonKind(l.Kind)})
	}

	tags := m.Tags
	if tags == nil {
		tags = []string{}
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, apigen.CoursePreview{
		Slug:           m.Slug,
		Title:          m.Title,
		Description:    m.Description,
		Category:       c.Category,
		Label:          apigen.CourseLabel(c.Label),
		Difficulty:     apigen.Difficulty(m.Difficulty),
		Tags:           tags,
		CoverURL:       "/api/v1/courses/" + url.PathEscape(m.Slug) + "/cover",
		Icon:           catalog.CategoryIcon(c.Category),
		IconURL:        m.IconURL,
		EstMinutes:     c.EstMinutes,
		LessonsCount:   len(lessons),
		LabsCount:      labs,
		Specialization: a.previewSpec(ctx, c.Spec),
		Lessons:        links,
	})
}

// previewSpec returns the course specialization, or nil when it is hidden.
func (a *API) previewSpec(ctx context.Context, spec string) *apigen.Specialization {
	s, err := a.Specs.Get(ctx, spec)
	if err != nil || !s.Published {
		return nil
	}
	return &apigen.Specialization{
		Slug:        s.Slug,
		Name:        s.Name,
		Icon:        s.Icon,
		IconURL:     s.IconURL,
		Description: s.Description,
		CoverURL:    "/api/v1/specializations/" + url.PathEscape(s.Slug) + "/cover",
	}
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
