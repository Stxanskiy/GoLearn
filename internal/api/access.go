package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/go-chi/chi/v5"
)

// courseLevel is the current user's relation to a course; higher levels include lower ones.
type courseLevel int

const (
	levelNone courseLevel = iota
	levelCoauthor
	levelOwner
	levelAdmin
)

// Permissions expressed as the minimum course level.
const (
	needEdit  = levelCoauthor
	needOwner = levelOwner
	needAdmin = levelAdmin
)

// courseLevel resolves what the user may do with a course; a draft inherits the access of its live course.
func (a *API) courseLevel(ctx context.Context, u *repository.User, m *model.Module) (courseLevel, error) {
	if m.DraftOf != nil && !u.IsAdmin() {
		live, err := a.Modules.GetByID(ctx, *m.DraftOf)
		if err != nil {
			return levelNone, err
		}
		m = live
	}
	switch {
	case u.IsAdmin():
		return levelAdmin, nil
	case !u.CanAuthor():
		return levelNone, nil
	case m.OwnerID != nil && *m.OwnerID == u.ID:
		return levelOwner, nil
	}
	ok, err := a.Authors.IsCoauthor(ctx, m.ID, u.ID)
	if err != nil || !ok {
		return levelNone, err
	}
	return levelCoauthor, nil
}

// canPreview reports whether the user may open a draft course or lesson.
func (a *API) canPreview(ctx context.Context, m *model.Module) (bool, error) {
	level, err := a.courseLevel(ctx, userFrom(ctx), m)
	return level >= needEdit, err
}

func courseAccess(level courseLevel) apigen.CourseAccess {
	name := apigen.CourseAccessLevelCoauthor
	switch level {
	case levelAdmin:
		name = apigen.CourseAccessLevelAdmin
	case levelOwner:
		name = apigen.CourseAccessLevelOwner
	}
	return apigen.CourseAccess{
		Level:            name,
		CanEdit:          level >= needEdit,
		CanUnpublish:     level >= needOwner,
		CanRequestReview: level >= needOwner,
		CanApprove:       level >= needAdmin,
		CanDelete:        level >= needOwner,
		CanManageAuthors: level >= needOwner,
		CanReorder:       level >= needAdmin,
	}
}

// checkEditable allows content changes only in drafts and unpublished courses that are not under review.
func (a *API) checkEditable(w http.ResponseWriter, r *http.Request, course courseRef) bool {
	if course.module.DraftOf == nil && course.module.Published {
		writeError(w, http.StatusConflict, codeDraftRequired, "published courses are edited through a draft")
		return false
	}
	pending, err := a.Reviews.Pending(r.Context(), course.live.ID)
	if err != nil {
		a.internalError(w, "admin: pending review", err)
		return false
	}
	if pending {
		writeError(w, http.StatusConflict, codeReviewPending, "course is under review")
		return false
	}
	return true
}

// checkCoursePublished allows owners to unpublish a live course without a draft; publishing needs a review.
func (a *API) checkCoursePublished(w http.ResponseWriter, r *http.Request, course courseRef, publish bool) bool {
	switch {
	case course.module.DraftOf != nil:
		writeError(w, http.StatusForbidden, codeForbidden, "drafts are never published")
		return false
	case publish:
		writeError(w, http.StatusForbidden, codeReviewRequired, "publishing requires an approved review")
		return false
	case course.level < needOwner:
		writeError(w, http.StatusForbidden, codeForbidden, "only the owner can unpublish the course")
		return false
	}
	_, err := a.Modules.DraftFor(r.Context(), course.module.ID)
	if err == nil {
		writeError(w, http.StatusConflict, codeDraftExists, "discard or submit the draft first")
		return false
	}
	if !isNotFound(err) {
		a.internalError(w, "admin: find draft", err)
		return false
	}
	return true
}

// checkLessonPublished lets owners hide lessons of a published course and toggle lessons of drafts and unpublished courses.
func (a *API) checkLessonPublished(w http.ResponseWriter, r *http.Request, course courseRef, publish bool) bool {
	if course.level < needOwner {
		writeError(w, http.StatusForbidden, codeForbidden, "only the owner can change lesson publication")
		return false
	}
	if course.module.DraftOf == nil && course.module.Published {
		if publish {
			writeError(w, http.StatusConflict, codeDraftRequired, "publish the lesson in the course draft")
			return false
		}
		return true
	}
	return a.checkEditable(w, r, course)
}

// courseRef is a course or draft the current user manages; live is the published course of a draft, otherwise module itself.
type courseRef struct {
	module *model.Module
	live   *model.Module
	level  courseLevel
}

// managedCourse loads a course and checks the user's level, writing 404/403/500 itself when it returns false.
func (a *API) managedCourse(w http.ResponseWriter, r *http.Request, id int, need courseLevel) (courseRef, bool) {
	ctx := r.Context()
	m, err := a.Modules.GetByID(ctx, id)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "course not found")
		return courseRef{}, false
	}
	if err != nil {
		a.internalError(w, "admin: load course", err)
		return courseRef{}, false
	}
	level, err := a.courseLevel(ctx, userFrom(ctx), m)
	if err != nil {
		a.internalError(w, "admin: course access", err)
		return courseRef{}, false
	}
	if level == levelNone {
		writeError(w, http.StatusNotFound, codeNotFound, "course not found")
		return courseRef{}, false
	}
	if level < need {
		writeError(w, http.StatusForbidden, codeForbidden, "not enough course permissions")
		return courseRef{}, false
	}
	live := m
	if m.DraftOf != nil {
		if live, err = a.Modules.GetByID(ctx, *m.DraftOf); err != nil {
			a.internalError(w, "admin: load live course", err)
			return courseRef{}, false
		}
	}
	return courseRef{module: m, live: live, level: level}, true
}

// managedCourseParam is managedCourse for the courseId path parameter.
func (a *API) managedCourseParam(w http.ResponseWriter, r *http.Request, need courseLevel) (courseRef, bool) {
	id, ok := pathID(w, r, "courseId", "course not found")
	if !ok {
		return courseRef{}, false
	}
	return a.managedCourse(w, r, id, need)
}

// managedLesson loads the lessonId lesson and checks access to its course.
func (a *API) managedLesson(w http.ResponseWriter, r *http.Request, need courseLevel) (*model.Lesson, courseRef, bool) {
	id, ok := pathID(w, r, "lessonId", "lesson not found")
	if !ok {
		return nil, courseRef{}, false
	}
	return a.managedLessonByID(w, r, id, need)
}

func (a *API) managedLessonByID(w http.ResponseWriter, r *http.Request, id int, need courseLevel) (*model.Lesson, courseRef, bool) {
	l, err := a.Lessons.GetByID(r.Context(), id)
	if err != nil {
		a.lessonLoadError(w, err)
		return nil, courseRef{}, false
	}
	course, ok := a.managedCourse(w, r, l.ModuleID, need)
	return l, course, ok
}

// pathID parses an integer path parameter, answering 404 when it is not a number.
func pathID(w http.ResponseWriter, r *http.Request, name, notFound string) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, name))
	if err != nil {
		writeError(w, http.StatusNotFound, codeNotFound, notFound)
		return 0, false
	}
	return id, true
}
