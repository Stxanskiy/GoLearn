package handler

import "net/http"

// TrainersPage lists the interactive tools (real runtimes) in one place.
func (h *Handler) TrainersPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "trainers", &struct{ PageTitle string }{PageTitle: "Тренажёры — TOT"})
}
