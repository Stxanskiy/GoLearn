package handler

import (
	"net/http"
)

// RoadmapModule is one course on the learning path.
type RoadmapModule struct {
	Num         int
	Title       string
	Slug        string
	Description string
	Category    string
	Icon        string
	LessonCount int
	Completed   int
	Pct         int
	Status      string // completed | in_progress | not_started
	Difficulty  string
	Duration    string
}

// RoadmapPath is one specialization rendered as an ordered route.
type RoadmapPath struct {
	Spec    Specialization
	Modules []RoadmapModule
	Done    int
	Total   int
	Pct     int
}

type RoadmapPageData struct {
	PageTitle string
	Paths     []RoadmapPath
	Active    string // slug of the path shown first
}

// RoadmapPage renders every specialization as a numbered route: the courses in
// the exact order the curriculum intends them to be taken, so it is obvious
// what comes next.
func (h *Handler) RoadmapPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("roadmap: get modules", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	allProgress, _ := h.progressRepo.GetAll(ctx, currentUserID(ctx))
	progressMap := make(map[int]string, len(allProgress))
	for _, p := range allProgress {
		progressMap[p.LessonID] = p.Status
	}

	// Group by specialization; modules already arrive in curriculum order.
	byspec := make(map[string][]RoadmapModule)
	for _, mod := range modules {
		spec := specForTrack(mod.Track)
		if spec == "gym" {
			continue // trainers are drills, not a course path
		}
		lessons, _ := h.lessonRepo.GetByModule(ctx, mod.ID)
		completed := 0
		for _, l := range lessons {
			if progressMap[l.ID] == "completed" {
				completed++
			}
		}
		status := "not_started"
		switch {
		case len(lessons) > 0 && completed == len(lessons):
			status = "completed"
		case completed > 0:
			status = "in_progress"
		}
		pct := 0
		if len(lessons) > 0 {
			pct = completed * 100 / len(lessons)
		}
		cat := mod.Category
		if cat == "" {
			cat = categorize(mod.Track, mod.Title, mod.Slug)
		}
		est := mod.EstMinutes
		if est == 0 {
			est = len(lessons) * 10
		}
		byspec[spec] = append(byspec[spec], RoadmapModule{
			Num:         len(byspec[spec]) + 1,
			Title:       mod.Title,
			Slug:        mod.Slug,
			Description: mod.Description,
			Category:    cat,
			Icon:        categoryIcon(cat),
			LessonCount: len(lessons),
			Completed:   completed,
			Pct:         pct,
			Status:      status,
			Difficulty:  mod.Difficulty,
			Duration:    humanDuration(est),
		})
	}

	specs, _ := h.specRepo.List(ctx)
	data := RoadmapPageData{PageTitle: "Дорожная карта — TOT"}
	for _, s := range specs {
		mods := byspec[s.Slug]
		if len(mods) == 0 {
			continue
		}
		path := RoadmapPath{
			Spec:    Specialization{Track: s.Slug, Name: s.Name, Icon: s.Icon, Desc: s.Description},
			Modules: mods,
			Total:   len(mods),
		}
		for _, m := range mods {
			if m.Status == "completed" {
				path.Done++
			}
		}
		if path.Total > 0 {
			path.Pct = path.Done * 100 / path.Total
		}
		data.Paths = append(data.Paths, path)
	}
	if len(data.Paths) > 0 {
		data.Active = data.Paths[0].Spec.Track
	}

	h.render(w, "roadmap", &data)
}
