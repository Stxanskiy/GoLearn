package handler

import "net/http"

type TrainersData struct {
	PageTitle string
	Gyms      []CourseCard
}

// TrainersPage lists interactive tools (real runtimes) and practice gyms.
func (h *Handler) TrainersPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	modules, _ := h.moduleRepo.GetAll(ctx)
	prog, _ := h.progressRepo.GetAll(ctx, currentUserID(ctx))
	pmap := make(map[int]string)
	for _, p := range prog {
		pmap[p.LessonID] = p.Status
	}
	var gyms []CourseCard
	for _, m := range modules {
		if m.Track == "gym" {
			gyms = append(gyms, h.buildCard(ctx, m, pmap))
		}
	}
	h.render(w, "trainers", &TrainersData{PageTitle: "Тренажёры — TOT", Gyms: gyms})
}
