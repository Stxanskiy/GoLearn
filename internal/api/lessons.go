package api

import (
	"context"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/content"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

const maxNotesLen = 20000

// lessonRef is a loaded lesson with its course.
type lessonRef struct {
	lesson *model.Lesson
	module *model.Module
}

// lessonByID loads a lesson visible to the current user, writing 404/500 itself when it returns false.
func (a *API) lessonByID(w http.ResponseWriter, r *http.Request) (lessonRef, bool) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "lessonId"))
	if err != nil {
		writeError(w, http.StatusNotFound, codeNotFound, "lesson not found")
		return lessonRef{}, false
	}
	l, err := a.Lessons.GetByID(ctx, id)
	if err != nil {
		a.lessonLoadError(w, err)
		return lessonRef{}, false
	}
	m, err := a.Modules.GetByID(ctx, l.ModuleID)
	if err != nil {
		a.lessonLoadError(w, err)
		return lessonRef{}, false
	}
	return a.visible(w, r, lessonRef{lesson: l, module: m})
}

// lessonBySlug loads a lesson by course and lesson slug with the same visibility rules as lessonByID.
func (a *API) lessonBySlug(w http.ResponseWriter, r *http.Request) (lessonRef, bool) {
	ctx := r.Context()
	m, err := a.Modules.GetBySlug(ctx, chi.URLParam(r, "courseSlug"))
	if err != nil {
		a.lessonLoadError(w, err)
		return lessonRef{}, false
	}
	l, err := a.Lessons.GetBySlug(ctx, m.ID, chi.URLParam(r, "lessonSlug"))
	if err != nil {
		a.lessonLoadError(w, err)
		return lessonRef{}, false
	}
	return a.visible(w, r, lessonRef{lesson: l, module: m})
}

func (a *API) visible(w http.ResponseWriter, r *http.Request, ref lessonRef) (lessonRef, bool) {
	if (!ref.lesson.Published || !ref.module.Published) && !userFrom(r.Context()).IsAdmin() {
		writeError(w, http.StatusNotFound, codeNotFound, "lesson not found")
		return lessonRef{}, false
	}
	return ref, true
}

func (a *API) lessonLoadError(w http.ResponseWriter, err error) {
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "lesson not found")
		return
	}
	a.internalError(w, "load lesson", err)
}

func (a *API) getLesson(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonBySlug(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	uid := userFrom(ctx).ID
	l, m := ref.lesson, ref.module

	siblings, err := a.Lessons.GetByModule(ctx, m.ID)
	if err != nil {
		a.internalError(w, "lesson: siblings", err)
		return
	}
	up, err := a.userProgress(ctx, uid)
	if err != nil {
		a.internalError(w, "lesson: progress", err)
		return
	}
	quiz, err := a.lessonQuiz(ctx, uid, l.ID)
	if err != nil {
		a.internalError(w, "lesson: quiz", err)
		return
	}
	_, tasks := a.Lessons.CountsForLesson(ctx, l.ID)
	attempts, err := a.QuizAttempts.Count(ctx, uid, l.ID)
	if err != nil {
		a.internalError(w, "lesson: attempts", err)
		return
	}

	out := apigen.LessonDetail{
		ID:          l.ID,
		Slug:        l.Slug,
		Title:       l.Title,
		Kind:        lessonKind(l.Kind),
		ContentHTML: content.Render(l.Format, l.Content),
		Nav:         lessonNav(*m, *l, siblings, up),
		Progress:    lessonProgress(*l, up, attempts),
		Quiz:        quiz,
		HasLab:      tasks > 0,
	}
	if s, ok := content.ExtractSQLSchema(l.Content); ok {
		out.SQLSchema = toSQLSchema(s)
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) visitLesson(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	if err := a.Progress.Start(ctx, userFrom(ctx).ID, ref.lesson.ID); err != nil {
		a.internalError(w, "visit lesson", err)
		return
	}
	a.writeLessonProgress(w, r, *ref.lesson)
}

func (a *API) completeLesson(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return
	}
	switch ref.lesson.Kind {
	case "quiz", "lab":
		writeError(w, http.StatusConflict, codeCompletionDerived, "quiz and lab lessons complete through their checks")
		return
	}
	ctx := r.Context()
	if err := a.Progress.Upsert(ctx, userFrom(ctx).ID, ref.lesson.ID, catalog.StatusCompleted); err != nil {
		a.internalError(w, "complete lesson", err)
		return
	}
	a.writeLessonProgress(w, r, *ref.lesson)
}

func (a *API) saveLessonNotes(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return
	}
	var body apigen.SaveLessonNotesJSONBody
	if !decodeJSON(w, r, &body) {
		return
	}
	if utf8.RuneCountInString(body.Notes) > maxNotesLen {
		writeValidation(w, map[string]string{"notes": fieldTooLong})
		return
	}
	ctx := r.Context()
	if err := a.Progress.SaveNotes(ctx, userFrom(ctx).ID, ref.lesson.ID, body.Notes); err != nil {
		a.internalError(w, "save notes", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeLessonProgress responds with the lesson's current derived progress.
func (a *API) writeLessonProgress(w http.ResponseWriter, r *http.Request, l model.Lesson) {
	ctx := r.Context()
	uid := userFrom(ctx).ID
	up, err := a.userProgress(ctx, uid)
	if err != nil {
		a.internalError(w, "lesson progress", err)
		return
	}
	attempts, err := a.QuizAttempts.Count(ctx, uid, l.ID)
	if err != nil {
		a.internalError(w, "lesson progress: attempts", err)
		return
	}
	writeJSON(w, http.StatusOK, lessonProgress(l, up, attempts))
}

func lessonProgress(l model.Lesson, up catalog.UserProgress, attempts int) apigen.LessonProgress {
	p := up.Lessons[l.ID]
	return apigen.LessonProgress{
		Status:       apigen.ProgressStatus(up.LessonStatus(l)),
		QuizAttempts: attempts,
		QuizScore:    p.QuizScore,
		QuizTotal:    p.QuizTotal,
		Notes:        p.Notes,
		CompletedAt:  p.CompletedAt,
	}
}

// lessonNav places a lesson among the course's published lessons; drafts get position 0 and no neighbours.
func lessonNav(m model.Module, l model.Lesson, siblings []model.Lesson, up catalog.UserProgress) apigen.LessonNav {
	c := catalog.BuildCourse(m, siblings, up)
	nav := apigen.LessonNav{
		Course:            apigen.LinkRef{Slug: m.Slug, Title: m.Title},
		Total:             len(siblings),
		CourseProgressPct: c.Pct,
	}
	for i, s := range siblings {
		if s.ID != l.ID {
			continue
		}
		nav.Position = i + 1
		if i > 0 {
			nav.Prev = lessonLink(siblings[i-1])
		}
		if i < len(siblings)-1 {
			nav.Next = lessonLink(siblings[i+1])
		}
	}
	return nav
}

func lessonLink(l model.Lesson) *apigen.LessonLink {
	return &apigen.LessonLink{Slug: l.Slug, Title: l.Title, Kind: lessonKind(l.Kind)}
}

func toSQLSchema(s *content.SQLSchema) *apigen.SQLSchema {
	out := &apigen.SQLSchema{Title: s.Title, Ddl: s.DDL, ExpectedColumns: s.Fields, Tables: []apigen.SQLTable{}}
	if out.ExpectedColumns == nil {
		out.ExpectedColumns = []string{}
	}
	for _, t := range s.Tables {
		table := apigen.SQLTable{ID: t.ID, Columns: []apigen.SQLColumn{}}
		for _, c := range t.Cols {
			desc := c.Desc
			table.Columns = append(table.Columns, apigen.SQLColumn{Name: c.Name, Type: c.Type, IsKey: c.IsKey, Description: &desc})
		}
		out.Tables = append(out.Tables, table)
	}
	return out
}

// lessonQuiz returns the public quiz (no answers except already answered ones), or nil without questions.
func (a *API) lessonQuiz(ctx context.Context, userID, lessonID int) (*apigen.LessonQuiz, error) {
	questions, err := a.quizQuestions(ctx, lessonID)
	if err != nil || len(questions) == 0 {
		return nil, err
	}
	answers, err := a.QuizAnswers.ForLesson(ctx, userID, lessonID)
	if err != nil {
		return nil, err
	}
	out := &apigen.LessonQuiz{Questions: make([]apigen.QuizQuestionPublic, 0, len(questions))}
	for _, q := range questions {
		pub := apigen.QuizQuestionPublic{ID: q.ID, QuestionHTML: content.Sanitize(q.Question), OptionsHTML: sanitizeAll(q.Options)}
		if sel, ok := answers[q.ID]; ok {
			res := answerResult(q, sel, true)
			pub.Answer = &res
		}
		out.Questions = append(out.Questions, pub)
	}
	return out, nil
}

func sanitizeAll(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = content.Sanitize(s)
	}
	return out
}
