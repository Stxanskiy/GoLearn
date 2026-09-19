package api

import (
	"errors"
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/repository"
)

func (a *API) adminOpenCourseDraft(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	if course.module.DraftOf != nil || !course.module.Published {
		writeError(w, http.StatusConflict, codeDraftNotNeeded, "only published courses have drafts")
		return
	}
	ctx := r.Context()
	status := http.StatusOK
	draft, err := a.Modules.DraftFor(ctx, course.module.ID)
	if isNotFound(err) {
		var id int
		id, err = a.Drafts.Create(ctx, course.module.ID)
		if isUniqueViolation(err) {
			draft, err = a.Modules.DraftFor(ctx, course.module.ID)
		} else if err == nil {
			status = http.StatusCreated
			draft, err = a.Modules.GetByID(ctx, id)
		}
	}
	if err != nil {
		a.internalError(w, "admin: open draft", err)
		return
	}
	a.writeAdminCourse(w, r, status, courseRef{module: draft, live: course.module, level: course.level}, nil)
}

func (a *API) adminDiscardCourseDraft(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needOwner)
	if !ok {
		return
	}
	a.discardDraft(w, r, course)
}

func (a *API) adminGetDraftChanges(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	ch, err := a.Drafts.Changes(r.Context(), course.live.ID)
	if errors.Is(err, repository.ErrNoDraft) {
		writeError(w, http.StatusNotFound, codeNotFound, "course has no draft")
		return
	}
	if err != nil {
		a.internalError(w, "admin: draft changes", err)
		return
	}
	out := apigen.DraftChanges{CourseFields: nonNil(ch.CourseFields), Lessons: make([]apigen.LessonChange, 0, len(ch.Lessons))}
	for _, l := range ch.Lessons {
		out.Lessons = append(out.Lessons, apigen.LessonChange{
			DraftLessonID: l.DraftLessonID, LiveLessonID: l.LiveLessonID, Slug: l.Slug, Title: l.Title,
			Change: apigen.LessonChangeChange(l.Change), Fields: nonNil(l.Fields),
			Questions: itemChanges(l.Questions), Tasks: itemChanges(l.Tasks),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// discardDraft deletes the draft of the course (course may be the draft itself).
func (a *API) discardDraft(w http.ResponseWriter, r *http.Request, course courseRef) {
	err := a.Drafts.Discard(r.Context(), course.live.ID)
	if errors.Is(err, repository.ErrNoDraft) {
		writeError(w, http.StatusNotFound, codeNotFound, "course has no draft")
		return
	}
	if err != nil {
		a.internalError(w, "admin: discard draft", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func itemChanges(c repository.ItemChanges) apigen.ItemChanges {
	return apigen.ItemChanges{Added: c.Added, Removed: c.Removed, Modified: c.Modified}
}

// adminOpenLessonDraft opens the course draft (creating it when needed) and answers with this
// lesson's copy inside it, so the studio can keep the author on the same lesson.
func (a *API) adminOpenLessonDraft(w http.ResponseWriter, r *http.Request) {
	lesson, course, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	// A draft lesson, or a lesson of an unpublished course, is already editable.
	if course.module.DraftOf != nil || !course.module.Published {
		writeJSON(w, http.StatusOK, apigen.DraftLessonRef{CourseID: course.module.ID, LessonID: lesson.ID})
		return
	}
	ctx := r.Context()
	draft, err := a.Modules.DraftFor(ctx, course.module.ID)
	status := http.StatusOK
	if isNotFound(err) {
		var id int
		id, err = a.Drafts.Create(ctx, course.module.ID)
		if isUniqueViolation(err) {
			draft, err = a.Modules.DraftFor(ctx, course.module.ID)
		} else if err == nil {
			status = http.StatusCreated
			draft, err = a.Modules.GetByID(ctx, id)
		}
	}
	if err != nil {
		a.internalError(w, "admin: open lesson draft", err)
		return
	}
	copyID, err := a.Lessons.DraftCopyOf(ctx, draft.ID, lesson.ID)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "lesson is not part of the draft")
		return
	}
	if err != nil {
		a.internalError(w, "admin: draft lesson copy", err)
		return
	}
	writeJSON(w, status, apigen.DraftLessonRef{CourseID: draft.ID, LessonID: copyID})
}
