package api

import (
	"net/http"
	"strings"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/courseio"
	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
)

const maxImportBody = 16 << 20

// importPlan is a validated course document with its diff against the stored course.
type importPlan struct {
	tree   model.CourseTree
	diff   repository.CourseDiff
	course *courseRef // nil when the import creates the course
	issues []courseio.Issue
}

func (a *API) adminExportCourse(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	tree, err := a.CourseIO.Export(r.Context(), course.module.ID)
	if err != nil {
		a.internalError(w, "admin: export course", err)
		return
	}
	data, err := courseio.Marshal(courseio.FromTree(tree))
	if err != nil {
		a.internalError(w, "admin: marshal course", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+course.module.Slug+`.course.json"`)
	_, _ = w.Write(data)
}

func (a *API) adminPreviewImport(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.planImport(w, r)
	if !ok {
		return
	}
	d := plan.diff
	writeJSON(w, http.StatusOK, apigen.ImportPreview{
		Slug: d.Slug, Title: d.Title, Exists: d.Exists, LessonsCount: len(plan.tree.Lessons),
		NewLessons: lessonRefs(d.New), UpdatedLessons: lessonRefs(d.Updated), RemovedLessons: lessonRefs(d.Removed),
		LostSubmissions: d.LostSubmissions, LostProgress: d.LostProgress,
		Issues: importIssues(plan.issues), Blocked: courseio.HasErrors(plan.issues),
	})
}

func (a *API) adminApplyImport(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.planImport(w, r)
	if !ok {
		return
	}
	if courseio.HasErrors(plan.issues) {
		issues := importIssues(plan.issues)
		writeJSON(w, http.StatusUnprocessableEntity, apigen.Error{Error: apigen.ErrorBody{
			Code: codeImportBlocked, Message: "course document has errors", Details: &apigen.ErrorDetails{Issues: &issues},
		}})
		return
	}
	ctx := r.Context()
	if plan.course != nil && plan.course.level < needPublish && len(plan.diff.Removed) > 0 {
		lessons, err := a.Lessons.GetByModuleAll(ctx, plan.course.module.ID)
		if err != nil {
			a.internalError(w, "import: lessons", err)
			return
		}
		removed := map[string]bool{}
		for _, l := range plan.diff.Removed {
			removed[l.Slug] = true
		}
		for _, l := range lessons {
			if l.Published && removed[l.Slug] {
				writeError(w, http.StatusForbidden, codeForbidden, "only the owner can remove published lessons")
				return
			}
		}
	}
	_, err := a.CourseIO.Upsert(ctx, plan.tree)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeSlugTaken, "course slug is taken")
		return
	}
	if err != nil {
		a.internalError(w, "import: upsert", err)
		return
	}
	m, err := a.Modules.GetBySlug(ctx, plan.tree.Module.Slug)
	if err != nil {
		a.internalError(w, "import: reload course", err)
		return
	}
	writeJSON(w, http.StatusOK, apigen.ImportResult{CourseID: m.ID, Created: plan.course == nil})
}

// planImport decodes and validates a course document, writing the error response itself when it returns false.
func (a *API) planImport(w http.ResponseWriter, r *http.Request) (importPlan, bool) {
	var doc courseio.Course
	if !decodeJSONLimit(w, r, &doc, maxImportBody) {
		return importPlan{}, false
	}
	ctx := r.Context()
	user := userFrom(ctx)
	var plan importPlan
	var cur *model.Module
	m, err := a.Modules.GetBySlug(ctx, strings.TrimSpace(doc.Slug))
	switch {
	case err == nil:
		level, err := a.courseLevel(ctx, user, m)
		if err != nil {
			a.internalError(w, "import: course access", err)
			return importPlan{}, false
		}
		if level < needEdit {
			writeError(w, http.StatusConflict, codeSlugTaken, "course slug is taken")
			return importPlan{}, false
		}
		cur, plan.course = m, &courseRef{module: m, level: level}
	case !isNotFound(err):
		a.internalError(w, "import: find course", err)
		return importPlan{}, false
	}

	input := apigen.AdminCourseInput{
		Slug: doc.Slug, Title: doc.Title, Track: doc.Track, Difficulty: apigen.Difficulty(doc.Difficulty),
		Description: &doc.Description, Category: &doc.Category, Accent: &doc.Accent, EstMinutes: &doc.EstMinutes, Tags: &doc.Tags,
	}
	if code := importLabel(doc.Label); code != "" {
		label := apigen.CourseLabel(code)
		input.Label = &label
	}
	mod, fields, err := a.courseFromInput(ctx, input, cur)
	if err != nil {
		a.internalError(w, "import: validate course", err)
		return importPlan{}, false
	}
	if c := doc.CoverImage; c != "" && !isHTTPURL(c) && !strings.HasPrefix(c, "data:image/") {
		fields["cover_image"] = fieldInvalidFormat
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return importPlan{}, false
	}

	plan.issues = courseio.Validate(doc)
	plan.tree = doc.ToTree()
	mod.CoverImage = doc.CoverImage
	if cur == nil {
		if mod.OrderNum, err = a.Modules.NextOrder(ctx); err != nil {
			a.internalError(w, "import: course order", err)
			return importPlan{}, false
		}
		mod.OwnerID = &user.ID
	} else {
		mod.OrderNum = cur.OrderNum
	}
	plan.tree.Module = mod
	publishLessons := plan.course == nil || plan.course.level >= needPublish
	for i := range plan.tree.Lessons {
		plan.tree.Lessons[i].Lesson.Published = publishLessons
	}
	if plan.diff, err = a.CourseIO.Diff(ctx, plan.tree); err != nil {
		a.internalError(w, "import: diff", err)
		return importPlan{}, false
	}
	return plan, true
}

// importLabel maps stored and localized label names to label codes; unknown values pass through for validation.
func importLabel(label string) string {
	switch label {
	case "Старт":
		return string(apigen.Start)
	case "Практика":
		return string(apigen.Practice)
	case "Вызов":
		return string(apigen.Challenge)
	}
	return label
}

func lessonRefs(refs []repository.LessonRef) []apigen.LinkRef {
	out := make([]apigen.LinkRef, 0, len(refs))
	for _, l := range refs {
		out = append(out, apigen.LinkRef{Slug: l.Slug, Title: l.Title})
	}
	return out
}

func importIssues(issues []courseio.Issue) []apigen.ImportIssue {
	out := make([]apigen.ImportIssue, 0, len(issues))
	for _, is := range issues {
		item := apigen.ImportIssue{Level: apigen.ImportIssueLevel(is.Level), Code: is.Code, Message: is.Msg}
		if is.LessonSlug != "" {
			item.LessonSlug = &is.LessonSlug
		}
		if is.Path != "" {
			item.Path = &is.Path
		}
		out = append(out, item)
	}
	return out
}
