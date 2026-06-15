package handler

import (
	"net/http"
)

type DashboardData struct {
	PageTitle        string
	SharedModules    []ModuleWithProgress
	BackendModules   []ModuleWithProgress
	DevopsModules    []ModuleWithProgress
	SecurityModules  []ModuleWithProgress
	TotalLessons     int
	CompletedCount   int
	InProgressCount  int
	AvgQuizScore     float64
	UserName         string
}

type ModuleWithProgress struct {
	ID          int
	Slug        string
	Title       string
	Description string
	Track       string
	Lessons     []LessonWithProgress
	Completed   int
	Total       int
}

type LessonWithProgress struct {
	ID       int
	Slug     string
	Title    string
	OrderNum int
	Status   string // "not_started", "in_progress", "completed"
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("get modules", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	allProgress, err := h.progressRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("get progress", "error", err)
		// Continue without progress data
	}

	progressMap := make(map[int]string)
	for _, p := range allProgress {
		progressMap[p.LessonID] = p.Status
	}

	stats, err := h.progressRepo.GetStats(ctx)
	if err != nil {
		h.log.Error("get stats", "error", err)
	}

	var data DashboardData
	data.PageTitle = "TOT — Dashboard"
	if stats != nil {
		data.TotalLessons = stats.TotalLessons
		data.CompletedCount = stats.CompletedCount
		data.InProgressCount = stats.InProgressCount
		data.AvgQuizScore = stats.AvgQuizScore
	}

	for _, mod := range modules {
		lessons, err := h.lessonRepo.GetByModule(ctx, mod.ID)
		if err != nil {
			h.log.Error("get lessons for module", "module", mod.Slug, "error", err)
			continue
		}

		mwp := ModuleWithProgress{
			ID:          mod.ID,
			Slug:        mod.Slug,
			Title:       mod.Title,
			Description: mod.Description,
			Track:       mod.Track,
			Total:       len(lessons),
		}

		for _, l := range lessons {
			status := "not_started"
			if s, ok := progressMap[l.ID]; ok {
				status = s
			}
			if status == "completed" {
				mwp.Completed++
			}
			mwp.Lessons = append(mwp.Lessons, LessonWithProgress{
				ID:       l.ID,
				Slug:     l.Slug,
				Title:    l.Title,
				OrderNum: l.OrderNum,
				Status:   status,
			})
		}

		switch mod.Track {
		case "shared":
			data.SharedModules = append(data.SharedModules, mwp)
		case "backend":
			data.BackendModules = append(data.BackendModules, mwp)
		case "devops":
			data.DevopsModules = append(data.DevopsModules, mwp)
		case "security-offense", "security-defense":
			data.SecurityModules = append(data.SecurityModules, mwp)
		default:
			data.SharedModules = append(data.SharedModules, mwp)
		}
	}

	h.render(w, "dashboard", &data)
}
