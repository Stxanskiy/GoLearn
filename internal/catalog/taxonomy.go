// Package catalog holds course taxonomy, progress status rules and cover rendering shared by HTML pages and the API.
package catalog

import (
	"strings"

	"github.com/backendraz/golearn/internal/model"
)

// GymSpec is the pseudo-specialization of trainer courses (not shown in the catalog).
const GymSpec = "gym"

// Course label codes; display text is localized by the client.
const (
	LabelStart     = "start"
	LabelPractice  = "practice"
	LabelChallenge = "challenge"
)

// SpecForTrack maps a module track to its specialization slug.
func SpecForTrack(track string) string {
	switch track {
	case "", "backend", "shared":
		return "devops"
	case "security-offense", "security-defense":
		return "security"
	default:
		return track
	}
}

// SpecTracks lists the module tracks that belong to one specialization.
func SpecTracks(spec string) []string {
	switch spec {
	case "devops":
		return []string{"devops", "", "backend", "shared"}
	case "security":
		return []string{"security", "security-offense", "security-defense"}
	default:
		return []string{spec}
	}
}

// Category returns the module's explicit category or one derived from its track and title.
func Category(m model.Module) string {
	if m.Category != "" {
		return m.Category
	}
	return Categorize(m.Track, m.Title, m.Slug)
}

// Categorize derives a topic tag from the module's track, title and slug.
func Categorize(track, title, slug string) string {
	if track == "golang" {
		return "Golang"
	}
	t := strings.ToLower(title + " " + slug)
	has := func(words ...string) bool {
		for _, w := range words {
			if strings.Contains(t, w) {
				return true
			}
		}
		return false
	}
	switch {
	case has("kubernetes", "helm", "k8s"):
		return "Kubernetes"
	case has("docker"):
		return "Docker"
	case has("postgres", "database", "sql", "база данных"):
		return "Database"
	case has("linux"):
		return "Linux"
	case has("git"):
		return "Git"
	case has("nginx", "ansible", "grafana", "prometheus", "ci/cd", "cicd", "монитор", "devops", "websocket"):
		return "DevOps"
	}
	switch track {
	case "database":
		return "Database"
	case "security", "security-offense", "security-defense":
		return "Security"
	default:
		return "DevOps"
	}
}

// CategoryIcon is the emoji shown for a category.
func CategoryIcon(cat string) string {
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

// LabelCode returns the course label code from the stored label or, if empty, from difficulty.
func LabelCode(m model.Module) string {
	switch m.Label {
	case "Старт", LabelStart:
		return LabelStart
	case "Практика", LabelPractice:
		return LabelPractice
	case "Вызов", LabelChallenge:
		return LabelChallenge
	}
	switch m.Difficulty {
	case "intermediate":
		return LabelPractice
	case "advanced", "expert":
		return LabelChallenge
	default:
		return LabelStart
	}
}

// EstMinutes is the module's time estimate, defaulting to 10 minutes per lesson.
func EstMinutes(m model.Module, lessons int) int {
	if m.EstMinutes > 0 {
		return m.EstMinutes
	}
	return lessons * 10
}
