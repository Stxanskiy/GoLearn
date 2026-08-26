package handler

import (
	"net/http"
	"time"

	"github.com/backendraz/golearn/internal/model"
)

type DashboardData struct {
	PageTitle       string
	UserName        string
	IsAdmin         bool
	ContinueURL     string
	ContinueTitle   string
	ContinueModule  string
	Overview        *model.ProgressOverview
	Weeks           [][]HeatCell
	MonthLabels     []MonthLabel
	DevopsModules   []ModuleWithProgress
	DatabaseModules []ModuleWithProgress
	SecurityModules []ModuleWithProgress
}

type HeatCell struct {
	Date  string
	Count int
	Level int
	Empty bool
}

type MonthLabel struct {
	Col  int
	Name string
}

type ModuleWithProgress struct {
	ID          int
	Slug        string
	Title       string
	Description string
	Track       string
	Category    string
	Icon        string
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

var ruMonths = []string{"Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"}

// buildHeatmap turns a date->count map into ~53 week columns (Mon..Sun rows).
func buildHeatmap(activity map[string]int) ([][]HeatCell, []MonthLabel) {
	end := time.Now()
	start := end.AddDate(0, 0, -363)
	// rewind start to Monday
	for int(start.Weekday()) != int(time.Monday) {
		start = start.AddDate(0, 0, -1)
	}

	var weeks [][]HeatCell
	var labels []MonthLabel
	cur := start
	lastMonth := -1
	for !cur.After(end) {
		week := make([]HeatCell, 0, 7)
		for d := 0; d < 7; d++ {
			if cur.After(end) {
				week = append(week, HeatCell{Empty: true})
			} else {
				key := cur.Format("2006-01-02")
				c := activity[key]
				week = append(week, HeatCell{Date: key, Count: c, Level: heatLevel(c)})
			}
			cur = cur.AddDate(0, 0, 1)
		}
		// month label when the first (Monday) of the week starts a new month
		mon := start.AddDate(0, 0, len(weeks)*7)
		if int(mon.Month())-1 != lastMonth {
			lastMonth = int(mon.Month()) - 1
			labels = append(labels, MonthLabel{Col: len(weeks), Name: ruMonths[lastMonth]})
		}
		weeks = append(weeks, week)
	}
	return weeks, labels
}

func heatLevel(c int) int {
	switch {
	case c <= 0:
		return 0
	case c <= 2:
		return 1
	case c <= 5:
		return 2
	case c <= 9:
		return 3
	default:
		return 4
	}
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data DashboardData
	data.PageTitle = "Мой прогресс — TOT"
	if u := GetUser(ctx); u != nil {
		data.UserName = u.Name
		data.IsAdmin = u.IsAdmin()
	}

	uid := currentUserID(ctx)
	overview, err := h.progressRepo.Overview(ctx, uid)
	if err != nil {
		h.log.Error("overview", "error", err)
		overview = &model.ProgressOverview{Activity: map[string]int{}, SimulatorsTot: 4, TrainersTot: 3}
	}
	overview.SimulatorsTot = len(scenarios())
	data.Overview = overview
	data.Weeks, data.MonthLabels = buildHeatmap(overview.Activity)

	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("get modules", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	allProgress, _ := h.progressRepo.GetAll(ctx, uid)
	progressMap := make(map[int]string)
	var contLessonID int
	var contTime time.Time
	for _, p := range allProgress {
		progressMap[p.LessonID] = p.Status
		if p.Status == "in_progress" && p.UpdatedAt.After(contTime) {
			contTime = p.UpdatedAt
			contLessonID = p.LessonID
		}
	}
	if contLessonID > 0 {
		if l, err := h.lessonRepo.GetByID(ctx, contLessonID); err == nil {
			if m, err := h.moduleRepo.GetByID(ctx, l.ModuleID); err == nil {
				data.ContinueTitle = l.Title
				data.ContinueModule = m.Title
				data.ContinueURL = "/module/" + m.Slug + "/lesson/" + l.Slug
				if l.Kind == "lab" {
					data.ContinueURL += "/tasks"
				}
			}
		}
	}

	for _, mod := range modules {
		if mod.Track == "gym" {
			continue // trainers are shown on /trainers, not the dashboard
		}
		lessons, err := h.lessonRepo.GetByModule(ctx, mod.ID)
		if err != nil {
			continue
		}
		mwp := ModuleWithProgress{
			ID: mod.ID, Slug: mod.Slug, Title: mod.Title, Description: mod.Description,
			Track: mod.Track, Total: len(lessons),
		}
		mwp.Category = mod.Category
		if mwp.Category == "" {
			mwp.Category = categorize(mod.Track, mod.Title, mod.Slug)
		}
		mwp.Icon = categoryIcon(mwp.Category)
		for _, l := range lessons {
			status := "not_started"
			if s, ok := progressMap[l.ID]; ok {
				status = s
			}
			if status == "completed" {
				mwp.Completed++
			}
			mwp.Lessons = append(mwp.Lessons, LessonWithProgress{
				ID: l.ID, Slug: l.Slug, Title: l.Title, OrderNum: l.OrderNum, Status: status,
			})
		}
		switch mod.Track {
		case "database":
			data.DatabaseModules = append(data.DatabaseModules, mwp)
		case "security", "security-offense", "security-defense":
			data.SecurityModules = append(data.SecurityModules, mwp)
		case "gym":
			// trainers have their own page
		default:
			data.DevopsModules = append(data.DevopsModules, mwp)
		}
	}

	h.render(w, "dashboard", &data)
}
