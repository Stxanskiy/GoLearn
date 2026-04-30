package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type statusRequest struct {
	Status string `json:"status"`
}

type notesRequest struct {
	Notes string `json:"notes"`
}

func (h *Handler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.Atoi(chi.URLParam(r, "lessonID"))
	if err != nil {
		http.Error(w, "Invalid lesson ID", 400)
		return
	}

	var req statusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", 400)
		return
	}

	switch req.Status {
	case "not_started", "in_progress", "completed":
	default:
		http.Error(w, "Invalid status", 400)
		return
	}

	if err := h.progressRepo.Upsert(r.Context(), lessonID, req.Status); err != nil {
		h.log.Error("update progress", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) SaveNotes(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.Atoi(chi.URLParam(r, "lessonID"))
	if err != nil {
		http.Error(w, "Invalid lesson ID", 400)
		return
	}

	var req notesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", 400)
		return
	}

	if err := h.progressRepo.SaveNotes(r.Context(), lessonID, req.Notes); err != nil {
		h.log.Error("save notes", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
