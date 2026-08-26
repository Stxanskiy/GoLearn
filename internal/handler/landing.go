package handler

import (
	"context"
	"net/http"

	"github.com/backendraz/golearn/internal/repository"
)

// LandingTrack is one specialization advertised on the public page.
type LandingTrack struct {
	Slug    string
	Name    string
	Icon    string
	Desc    string
	Courses int
}

type LandingData struct {
	PageTitle  string
	Stats      repository.PlatformStats
	Tracks     []LandingTrack
	FirstCours string // slug of the course the "start here" button points at
}

// Home serves the public landing page to visitors and the dashboard to anyone
// already signed in, so "/" is the right destination in both states.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil && cookie.Value != "" {
		if user, err := h.userRepo.GetUserBySession(r.Context(), cookie.Value); err == nil {
			h.Dashboard(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
			return
		}
	}
	h.LandingPage(w, r)
}

// LandingPage renders what the platform is, using the real course counts rather
// than claims that could drift away from the content.
func (h *Handler) LandingPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.moduleRepo.Stats(ctx)
	if err != nil {
		h.log.Error("landing: stats", "error", err)
	}

	data := LandingData{PageTitle: "TOT — практический курс DevOps", Stats: stats}

	modules, err := h.moduleRepo.GetAll(ctx)
	if err != nil {
		h.log.Error("landing: modules", "error", err)
	}
	perSpec := make(map[string]int)
	for _, m := range modules {
		spec := specForTrack(m.Track)
		perSpec[spec]++
		if spec == "devops" && data.FirstCours == "" {
			data.FirstCours = m.Slug // modules arrive in curriculum order
		}
	}

	specs, _ := h.specRepo.List(ctx)
	for _, s := range specs {
		if perSpec[s.Slug] == 0 {
			continue
		}
		data.Tracks = append(data.Tracks, LandingTrack{
			Slug: s.Slug, Name: s.Name, Icon: s.Icon, Desc: s.Description, Courses: perSpec[s.Slug],
		})
	}

	h.render(w, "landing", &data)
}
