package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type CourseCard struct {
	Title      string
	Slug       string
	Category   string
	Icon       string
	Label      string // Старт | Практика | Вызов
	Cover      string // image URL (real photo or generated SVG)
	Tags       []string
	Difficulty string
	Chapters   int
	Completed  int
	Pct        int
	Duration   string // human-readable estimate
	EstMinutes int    // raw estimate, used for sorting
	Started    bool
	Num        int    // position in the specialization's curriculum
	Status     string // completed | in_progress | not_started
}

type CoursesData struct {
	PageTitle string
	Specs     []Specialization
	Total     int
	Done      int
}

// Specialization is a top-level track shown as a card that expands to its courses.
type Specialization struct {
	Track string
	Name  string
	Icon  string
	Desc  string
	Count int
	Done  int
	Cards []CourseCard
}

func humanDuration(minutes int) string {
	if minutes <= 0 {
		return ""
	}
	if minutes < 60 {
		return fmt.Sprintf("~%d мин", minutes)
	}
	h, m := minutes/60, minutes%60
	if m == 0 {
		return fmt.Sprintf("~%d ч", h)
	}
	return fmt.Sprintf("~%d ч %d мин", h, m)
}

// buildCard turns a module into a catalog card using the progress map.
func (h *Handler) buildCard(ctx context.Context, m model.Module, pmap map[int]string) CourseCard {
	lessons, _ := h.lessonRepo.GetByModule(ctx, m.ID)
	done, started := 0, false
	for _, l := range lessons {
		switch pmap[l.ID] {
		case "completed":
			done++
			started = true
		case "in_progress":
			started = true
		}
	}
	pct := 0
	if len(lessons) > 0 {
		pct = done * 100 / len(lessons)
	}
	cat := catalog.Category(m)
	label := m.Label
	if label == "" {
		label = deriveLabel(m.Difficulty)
	}
	est := m.EstMinutes
	if est == 0 {
		est = len(lessons) * 10
	}
	status := "not_started"
	switch {
	case pct == 100 && len(lessons) > 0:
		status = "completed"
	case started:
		status = "in_progress"
	}
	return CourseCard{
		Title: m.Title, Slug: m.Slug, Category: cat, Icon: catalog.CategoryIcon(cat),
		Label: label, Cover: "/api/courses/" + m.Slug + "/cover", Tags: m.Tags, Difficulty: m.Difficulty,
		Chapters: len(lessons), Completed: done, Pct: pct,
		Duration: humanDuration(est), EstMinutes: est, Started: started, Status: status,
	}
}

// CoursesPage renders the top-level specialization cards.
func (h *Handler) CoursesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("courses: get modules", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}
	prog, _ := h.progressRepo.GetAll(ctx, currentUserID(ctx))
	pmap := make(map[int]string)
	for _, p := range prog {
		pmap[p.LessonID] = p.Status
	}

	count := make(map[string]int)
	done := make(map[string]int)
	total, doneCourses := 0, 0
	for _, m := range modules {
		if m.IsTrainer {
			continue // trainers live on /trainers, not in the catalog
		}
		sp := catalog.SpecForTrack(m.Track)
		count[sp]++
		total++
		c := h.buildCard(ctx, m, pmap)
		if c.Pct == 100 {
			done[sp]++
			doneCourses++
		}
	}

	dbSpecs, _ := h.specRepo.ListPublished(ctx)
	var specs []Specialization
	for _, s := range dbSpecs {
		specs = append(specs, Specialization{
			Track: s.Slug, Name: s.Name, Icon: s.Icon, Desc: s.Description,
			Count: count[s.Slug], Done: done[s.Slug],
		})
	}

	h.render(w, "courses", &CoursesData{
		PageTitle: "Курсы — TOT",
		Specs:     specs,
		Total:     total,
		Done:      doneCourses,
	})
}

type SectionData struct {
	PageTitle  string
	Spec       Specialization
	Cards      []CourseCard
	Categories []CategoryFilter
}

type CategoryFilter struct {
	Name  string
	Count int
}

// SectionPage renders one specialization: its courses + category filter.
func (h *Handler) SectionPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	track := chi.URLParam(r, "track")

	s, err := h.specRepo.Get(ctx, track)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	spec := Specialization{Track: s.Slug, Name: s.Name, Icon: s.Icon, Desc: s.Description}

	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		http.Error(w, "Internal error", 500)
		return
	}
	prog, _ := h.progressRepo.GetAll(ctx, currentUserID(ctx))
	pmap := make(map[int]string)
	for _, p := range prog {
		pmap[p.LessonID] = p.Status
	}

	catCount := make(map[string]int)
	var cards []CourseCard
	for _, m := range modules {
		if catalog.SpecForTrack(m.Track) != track {
			continue
		}
		c := h.buildCard(ctx, m, pmap)
		c.Num = len(cards) + 1 // modules arrive in curriculum order
		cards = append(cards, c)
		catCount[c.Category]++
	}

	order := []string{"DevOps", "Linux", "Docker", "Kubernetes", "Git", "Database", "Golang", "Security"}
	var cats []CategoryFilter
	for _, c := range order {
		if catCount[c] > 0 {
			cats = append(cats, CategoryFilter{Name: c, Count: catCount[c]})
		}
	}

	h.render(w, "section", &SectionData{
		PageTitle:  spec.Name + " — TOT",
		Spec:       spec,
		Cards:      cards,
		Categories: cats,
	})
}
