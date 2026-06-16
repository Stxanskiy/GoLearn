package handler

import (
	"net/http"
)

type RoadmapTrack struct {
	Name        string
	Slug        string
	Description string
	Levels      []RoadmapLevel
}

type RoadmapLevel struct {
	Name    string
	Modules []RoadmapModule
}

type RoadmapModule struct {
	Title       string
	Slug        string
	Description string
	LessonCount int
	Completed   int
	Status      string
	Track       string
	Difficulty  string
}

type RoadmapPageData struct {
	PageTitle    string
	SharedBase   []RoadmapLevel
	BackendTrack []RoadmapLevel
	DevopsTrack  []RoadmapLevel
}

func (h *Handler) RoadmapPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("get modules", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	allProgress, _ := h.progressRepo.GetAll(ctx)
	progressMap := make(map[int]string)
	for _, p := range allProgress {
		progressMap[p.LessonID] = p.Status
	}

	difficultyToLevel := map[string]int{
		"beginner":     0,
		"intermediate": 1,
		"advanced":     2,
		"expert":       3,
	}

	sharedLevels := []RoadmapLevel{
		{Name: "Beginner — Основы языка"},
		{Name: "Intermediate — Продвинутые основы"},
	}
	backendLevels := []RoadmapLevel{
		{Name: "Intermediate — Веб и данные"},
		{Name: "Advanced — Архитектура и качество"},
		{Name: "Expert — Производительность"},
	}
	devopsLevels := []RoadmapLevel{
		{Name: "Intermediate — Контейнеризация"},
		{Name: "Advanced — CI/CD и мониторинг"},
		{Name: "Expert — Оркестрация"},
	}

	for _, mod := range modules {
		lessons, _ := h.lessonRepo.GetByModule(ctx, mod.ID)

		completed := 0
		for _, l := range lessons {
			if s, ok := progressMap[l.ID]; ok && s == "completed" {
				completed++
			}
		}

		status := "not_started"
		if completed == len(lessons) && len(lessons) > 0 {
			status = "completed"
		} else if completed > 0 {
			status = "in_progress"
		}

		rm := RoadmapModule{
			Title:       mod.Title,
			Slug:        mod.Slug,
			Description: mod.Description,
			LessonCount: len(lessons),
			Completed:   completed,
			Status:      status,
			Track:       mod.Track,
			Difficulty:  mod.Difficulty,
		}

		level := difficultyToLevel[mod.Difficulty]

		switch mod.Track {
		case "golang", "shared", "backend":
			if level >= 0 && level < len(sharedLevels) {
				sharedLevels[level].Modules = append(sharedLevels[level].Modules, rm)
			} else {
				sharedLevels[len(sharedLevels)-1].Modules = append(sharedLevels[len(sharedLevels)-1].Modules, rm)
			}
		case "devops", "database":
			idx := level - 1
			if idx < 0 {
				idx = 0
			}
			if idx >= len(devopsLevels) {
				idx = len(devopsLevels) - 1
			}
			devopsLevels[idx].Modules = append(devopsLevels[idx].Modules, rm)
		}
	}

	data := RoadmapPageData{
		PageTitle:    "Learning Roadmap",
		SharedBase:   sharedLevels,
		BackendTrack: backendLevels,
		DevopsTrack:  devopsLevels,
	}

	h.render(w, "roadmap", &data)
}
