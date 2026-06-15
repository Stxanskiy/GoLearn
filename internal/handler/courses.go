package handler

import (
	"net/http"
	"strings"
)

type CourseCard struct {
	Title      string
	Slug       string
	Category   string
	Icon       string
	Difficulty string
	Lessons    int
	Completed  int
	Pct        int
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
	case "Security":
		return "🛡️"
	default:
		return "🚀"
	}
}

type CoursesData struct {
	PageTitle  string
	Cards      []CourseCard
	Categories []string
}

// categorize derives a topic tag from the module's track and title.
func categorize(track, title, slug string) string {
	t := strings.ToLower(title + " " + slug)
	switch {
	case strings.Contains(t, "kubernetes") || strings.Contains(t, "helm") || strings.Contains(t, "k8s"):
		return "Kubernetes"
	case strings.Contains(t, "docker"):
		return "Docker"
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
	case "backend":
		return "Backend"
	case "security-offense", "security-defense":
		return "Security"
	default:
		return "Основы"
	}
}

// CoursesPage renders a clean, filterable catalog of all courses.
func (h *Handler) CoursesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("courses: get modules", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}
	prog, _ := h.progressRepo.GetAll(ctx)
	pmap := make(map[int]string)
	for _, p := range prog {
		pmap[p.LessonID] = p.Status
	}

	catSet := make(map[string]bool)
	var cards []CourseCard
	for _, m := range modules {
		lessons, _ := h.lessonRepo.GetByModule(ctx, m.ID)
		done := 0
		for _, l := range lessons {
			if pmap[l.ID] == "completed" {
				done++
			}
		}
		pct := 0
		if len(lessons) > 0 {
			pct = done * 100 / len(lessons)
		}
		cat := categorize(m.Track, m.Title, m.Slug)
		catSet[cat] = true
		cards = append(cards, CourseCard{
			Title: m.Title, Slug: m.Slug, Category: cat, Icon: categoryIcon(cat), Difficulty: m.Difficulty,
			Lessons: len(lessons), Completed: done, Pct: pct,
		})
	}

	order := []string{"DevOps", "Linux", "Docker", "Kubernetes", "Git", "Backend", "Основы", "Security"}
	var cats []string
	for _, c := range order {
		if catSet[c] {
			cats = append(cats, c)
		}
	}

	h.render(w, "courses", &CoursesData{
		PageTitle:  "Курсы — TOT",
		Cards:      cards,
		Categories: cats,
	})
}
