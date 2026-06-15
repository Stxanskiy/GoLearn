package handler

import (
	"net/http"
)

func (h *Handler) PlaygroundPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
	}{
		PageTitle: "Go Playground",
	}
	h.render(w, "playground", data)
}
