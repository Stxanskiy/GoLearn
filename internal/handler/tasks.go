package handler

import (
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type TasksPageData struct {
	PageTitle   string
	Module      *model.Module
	Lesson      *model.Lesson
	Tasks       []model.Task
	IsShellLab  bool
	SessionTask int // shared sandbox session id for the whole lab (first task)
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
		PageTitle: "Лаборатория: " + lesson.Title,
		Module:    mod,
		Lesson:    lesson,
		Tasks:     tasks,
	}
	if len(tasks) > 0 && tasks[0].Kind == "shell" {
		data.IsShellLab = true
		data.SessionTask = tasks[0].ID
	}

	h.render(w, "tasks", &data)
}
