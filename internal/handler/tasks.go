package handler

import (
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type TasksPageData struct {
	PageTitle string
	Module    *model.Module
	Lesson    *model.Lesson
	Tasks     []model.Task
}

func (h *Handler) TasksPage(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := h.lessonRepo.GetTasks(ctx, lesson.ID)
	if err != nil {
		h.log.Error("get tasks", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	data := TasksPageData{
		PageTitle: "Tasks: " + lesson.Title,
		Module:    mod,
		Lesson:    lesson,
		Tasks:     tasks,
	}

	h.render(w, "tasks", &data)
}
