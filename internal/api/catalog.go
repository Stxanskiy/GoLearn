package api

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// userProgress loads a user's raw progress and passed labs.
func (a *API) userProgress(ctx context.Context, userID int) (catalog.UserProgress, error) {
	rows, err := a.Progress.GetAll(ctx, userID)
	if err != nil {
		return catalog.UserProgress{}, err
	}
	labs, err := a.Submissions.LessonLabStatus(ctx, userID)
	if err != nil {
		return catalog.UserProgress{}, err
	}
	return catalog.NewUserProgress(rows, labs), nil
}

// publishedCourses builds every published course (trainers included) with the user's progress, in curriculum order.
func (a *API) publishedCourses(ctx context.Context, userID int) ([]catalog.Course, error) {
	modules, err := a.Modules.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	outline, err := a.Lessons.ListPublishedOutline(ctx)
	if err != nil {
		return nil, err
	}
	up, err := a.userProgress(ctx, userID)
	if err != nil {
		return nil, err
	}
	byModule := make(map[int][]model.Lesson)
	for _, l := range outline {
		byModule[l.ModuleID] = append(byModule[l.ModuleID], l)
	}
	courses := make([]catalog.Course, 0, len(modules))
	for _, m := range modules {
		courses = append(courses, catalog.BuildCourse(m, byModule[m.ID], up))
	}
	return courses, nil
}

func (a *API) getCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	courses, err := a.publishedCourses(ctx, userFrom(ctx).ID)
	if err != nil {
		a.internalError(w, "catalog: courses", err)
		return
	}
	specs, err := a.Specs.ListPublished(ctx)
	if err != nil {
		a.internalError(w, "catalog: specs", err)
		return
	}
	filter, ok := catalogFilterFrom(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, buildCatalog(specs, courses, filter))
}

// buildCatalog assembles the response: every published specialization is listed so the
// filter chips stay complete, while its courses and the trainers are narrowed by the filter.
func buildCatalog(specs []model.Specialization, courses []catalog.Course, filter catalogFilter) apigen.Catalog {
	shown := filter.keep(courses)
	out := apigen.Catalog{
		Specializations: make([]apigen.SpecializationWithCourses, 0, len(specs)),
		Trainers:        trainerCards(shown),
	}
	for _, s := range specs {
		out.Specializations = append(out.Specializations, specWithCourses(s, courses, shown))
	}
	return out
}

func (a *API) getSpecialization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	spec, err := a.Specs.Get(ctx, chi.URLParam(r, "specSlug"))
	if isNotFound(err) || (err == nil && !spec.Published && !userFrom(ctx).IsAdmin()) {
		writeError(w, http.StatusNotFound, codeNotFound, "specialization not found")
		return
	}
	if err != nil {
		a.internalError(w, "specialization: get", err)
		return
	}
	courses, err := a.publishedCourses(ctx, userFrom(ctx).ID)
	if err != nil {
		a.internalError(w, "specialization: courses", err)
		return
	}
	writeJSON(w, http.StatusOK, specWithCourses(*spec, courses, courses))
}

func (a *API) getCourse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := userFrom(ctx)
	m, err := a.Modules.GetBySlug(ctx, chi.URLParam(r, "courseSlug"))
	if err == nil && !m.Published {
		var ok bool
		if ok, err = a.canPreview(ctx, m); err == nil && !ok {
			err = pgx.ErrNoRows
		}
	}
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "course not found")
		return
	}
	if err != nil {
		a.internalError(w, "course: get", err)
		return
	}
	lessons, err := a.Lessons.GetByModule(ctx, m.ID)
	if err != nil {
		a.internalError(w, "course: lessons", err)
		return
	}
	up, err := a.userProgress(ctx, user.ID)
	if err != nil {
		a.internalError(w, "course: progress", err)
		return
	}
	// A trainer stands on its own, so it gets no place in the course path.
	var prev, next *model.Module
	if !m.IsTrainer {
		if prev, next, err = a.Modules.Neighbors(ctx, *m, catalog.SpecTracks(catalog.SpecForTrack(m.Track))); err != nil {
			a.internalError(w, "course: neighbors", err)
			return
		}
	}

	c := catalog.BuildCourse(*m, lessons, up)
	nextIdx := c.NextLesson()
	items := make([]apigen.CourseItem, 0, len(c.Lessons))
	for i, ls := range c.Lessons {
		items = append(items, apigen.CourseItem{
			LessonID: ls.Lesson.ID,
			Slug:     ls.Lesson.Slug,
			Title:    ls.Lesson.Title,
			Kind:     lessonKind(ls.Lesson.Kind),
			Status:   apigen.ProgressStatus(ls.Status),
			IsNext:   i == nextIdx,
		})
	}
	card := courseCard(c)
	card.Description = &m.Description
	writeJSON(w, http.StatusOK, apigen.CourseDetail{
		Course:         card,
		Items:          items,
		CompletedCount: c.Completed,
		ProgressPct:    c.Pct,
		PrevCourse:     linkRef(prev),
		NextCourse:     linkRef(next),
	})
}

// specWithCourses lists the courses of shown, while courses_done always counts the whole specialization.
func specWithCourses(s model.Specialization, all, shown []catalog.Course) apigen.SpecializationWithCourses {
	cards := courseCards(shown, s.Slug)
	done := 0
	for _, c := range all {
		if c.Spec == s.Slug && !c.Module.IsTrainer && c.Status == catalog.StatusCompleted {
			done++
		}
	}
	return apigen.SpecializationWithCourses{
		Slug:        s.Slug,
		Name:        s.Name,
		Icon:        s.Icon,
		IconURL:     s.IconURL,
		Description: s.Description,
		CoverURL:    "/api/v1/specializations/" + url.PathEscape(s.Slug) + "/cover",
		Courses:     cards,
		CoursesDone: done,
	}
}

// courseCards returns cards of the regular courses that belong to spec; trainers have their own list.
func courseCards(courses []catalog.Course, spec string) []apigen.CourseCard {
	return progressOrderedCards(courses, func(c catalog.Course) bool {
		return c.Spec == spec && !c.Module.IsTrainer
	})
}

// trainerCards returns cards of the practice-only courses, whatever specialization they belong to.
func trainerCards(courses []catalog.Course) []apigen.CourseCard {
	return progressOrderedCards(courses, func(c catalog.Course) bool { return c.Module.IsTrainer })
}

// progressOrderedCards selects courses and returns them in the catalogue order: started, untouched, completed.
func progressOrderedCards(courses []catalog.Course, keep func(catalog.Course) bool) []apigen.CourseCard {
	picked := []catalog.Course{}
	for _, c := range courses {
		if keep(c) {
			picked = append(picked, c)
		}
	}
	catalog.SortByProgress(picked)

	cards := make([]apigen.CourseCard, 0, len(picked))
	for _, c := range picked {
		cards = append(cards, courseCard(c))
	}
	return cards
}

func courseCard(c catalog.Course) apigen.CourseCard {
	tags := c.Module.Tags
	if tags == nil {
		tags = []string{}
	}
	return apigen.CourseCard{
		ID:               c.Module.ID,
		Slug:             c.Module.Slug,
		Title:            c.Module.Title,
		SpecSlug:         c.Spec,
		Category:         c.Category,
		Label:            apigen.CourseLabel(c.Label),
		Difficulty:       apigen.Difficulty(c.Module.Difficulty),
		IsTrainer:        c.Module.IsTrainer,
		Tags:             tags,
		CoverURL:         "/api/v1/courses/" + url.PathEscape(c.Module.Slug) + "/cover",
		Icon:             catalog.CategoryIcon(c.Category),
		IconURL:          c.Module.IconURL,
		LessonsCount:     len(c.Lessons),
		LessonsCompleted: c.Completed,
		ProgressPct:      c.Pct,
		EstMinutes:       c.EstMinutes,
		Status:           apigen.ProgressStatus(c.Status),
		LastActivity:     nilTime(c.LastActivity),
	}
}

// nilTime maps the zero activity timestamp of an untouched course to JSON null.
func nilTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func lessonKind(kind string) apigen.LessonKind {
	if kind == "" {
		return apigen.LessonKindTheory
	}
	return apigen.LessonKind(kind)
}

func linkRef(m *model.Module) *apigen.LinkRef {
	if m == nil {
		return nil
	}
	return &apigen.LinkRef{Slug: m.Slug, Title: m.Title}
}
