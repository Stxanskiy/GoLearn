package handler

import (
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type ModulePageData struct {
	PageTitle string
	Module    *model.Module
	Lessons   []LessonWithProgress
}

type LessonPageData struct {
	PageTitle   string
	Module      *model.Module
	Lesson      *model.Lesson
	ContentHTML string
	Progress    *model.Progress
	HasQuiz     bool
	HasTasks    bool
	PrevLesson  *model.Lesson
	NextLesson  *model.Lesson
}

func (h *Handler) ModulePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lessons, err := h.lessonRepo.GetByModule(ctx, mod.ID)
	if err != nil {
		h.log.Error("get lessons", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	allProgress, _ := h.progressRepo.GetAll(ctx)
	progressMap := make(map[int]string)
	for _, p := range allProgress {
		progressMap[p.LessonID] = p.Status
	}

	data := ModulePageData{
		PageTitle: mod.Title,
		Module:    mod,
	}
	for _, l := range lessons {
		status := "not_started"
		if s, ok := progressMap[l.ID]; ok {
			status = s
		}
		data.Lessons = append(data.Lessons, LessonWithProgress{
			ID: l.ID, Slug: l.Slug, Title: l.Title, OrderNum: l.OrderNum, Status: status,
		})
	}

	h.render(w, "module", &data)
}

func (h *Handler) LessonPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")
	lessonSlug := chi.URLParam(r, "lessonSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lesson, err := h.lessonRepo.GetBySlug(ctx, mod.ID, lessonSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Mark as in_progress if not started
	progress, _ := h.progressRepo.Get(ctx, lesson.ID)
	if progress == nil {
		_ = h.progressRepo.Upsert(ctx, lesson.ID, "in_progress")
		progress, _ = h.progressRepo.Get(ctx, lesson.ID)
	}

	// Check if quiz/tasks exist
	_, questions, _ := h.lessonRepo.GetQuiz(ctx, lesson.ID)
	tasks, _ := h.lessonRepo.GetTasks(ctx, lesson.ID)

	// Find prev/next lessons
	allLessons, _ := h.lessonRepo.GetByModule(ctx, mod.ID)
	var prev, next *model.Lesson
	for i, l := range allLessons {
		if l.ID == lesson.ID {
			if i > 0 {
				prev = &allLessons[i-1]
			}
			if i < len(allLessons)-1 {
				next = &allLessons[i+1]
			}
			break
		}
	}

	data := LessonPageData{
		PageTitle:   lesson.Title,
		Module:      mod,
		Lesson:      lesson,
		ContentHTML: lesson.Content,
		Progress:    progress,
		HasQuiz:     len(questions) > 0,
		HasTasks:    len(tasks) > 0,
		PrevLesson:  prev,
		NextLesson:  next,
	}

	h.render(w, "lesson", &data)
}
