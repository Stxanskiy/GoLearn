package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

func categoryIcon(cat string) string {
	switch cat {
	case "Linux":
		return "🐧"
	case "Docker":
		return "🐳"
	case "Kubernetes":
		return "☸️"
	case "Git":
		return "🌿"
	case "DevOps":
		return "♾️"
	case "Backend":
		return "⚙️"
	case "Golang":
		return "🐹"
	case "Database":
		return "🗄️"
	case "Security":
		return "🛡️"
	default:
		return "🚀"
	}
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

func specForTrack(track string) string {
	switch track {
	case "devops":
		return "devops"
	case "database":
		return "database"
	case "gym":
		return "gym" // trainers — not a catalog specialization
	case "security", "security-offense", "security-defense":
		return "security"
	default:
		return "devops"
	}
}

// specTracks lists the module tracks that belong to one specialization — used
// to walk the curriculum inside a single learning path.
func specTracks(spec string) []string {
	switch spec {
	case "devops":
		return []string{"devops"}
	case "database":
		return []string{"database"}
	case "gym":
		return []string{"gym"}
	case "security":
		return []string{"security", "security-offense", "security-defense"}
	default:
		return []string{"devops"}
	}
}

// categorize derives a topic tag from the module's track and title.
func categorize(track, title, slug string) string {
	if track == "golang" {
		return "Golang"
	}
	t := strings.ToLower(title + " " + slug)
	switch {
	case strings.Contains(t, "kubernetes") || strings.Contains(t, "helm") || strings.Contains(t, "k8s"):
		return "Kubernetes"
	case strings.Contains(t, "docker"):
		return "Docker"
	case strings.Contains(t, "postgres") || strings.Contains(t, "database") || strings.Contains(t, "sql") || strings.Contains(t, "база данных"):
		return "Database"
	case strings.Contains(t, "linux"):
		return "Linux"
	case strings.Contains(t, "git"):
		return "Git"
	case strings.Contains(t, "nginx") || strings.Contains(t, "ansible") || strings.Contains(t, "grafana") ||
		strings.Contains(t, "prometheus") || strings.Contains(t, "ci/cd") || strings.Contains(t, "cicd") ||
		strings.Contains(t, "монитор") || strings.Contains(t, "devops") || strings.Contains(t, "websocket"):
		return "DevOps"
	}
	switch track {
	case "database":
		return "Database"
	case "security", "security-offense", "security-defense":
		return "Security"
	case "golang":
		return "Golang"
	default:
		return "DevOps"
	}
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
	cat := m.Category
	if cat == "" {
		cat = categorize(m.Track, m.Title, m.Slug)
	}
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
		Title: m.Title, Slug: m.Slug, Category: cat, Icon: categoryIcon(cat),
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
		sp := specForTrack(m.Track)
		if sp == "gym" {
			continue // trainers live on /trainers, not in the catalog
		}
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
		if specForTrack(m.Track) != track {
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
