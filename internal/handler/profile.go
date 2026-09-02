package handler

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/backendraz/golearn/internal/repository"
)

type ProfileData struct {
	PageTitle string
	User      *repository.User
	Initial   string
	Joined    string
	Completed int
	Started   int
	RoleLabel string
}

// firstInitial returns the first letter (rune-safe, uppercased) of name or email.
func firstInitial(name, email string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		s = email
	}
	for _, r := range s {
		return string(unicode.ToUpper(r))
	}
	return "?"
}

// ProfilePage shows the signed-in user's account: identity, join date and a
// short learning summary. (Subscription/tariff info lands here once payments are
// wired.)
func (h *Handler) ProfilePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUser(ctx)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	prog, _ := h.progressRepo.GetAll(ctx, user.ID)
	completed := 0
	for _, p := range prog {
		if p.Status == "completed" {
			completed++
		}
	}

	role := "Ученик"
	if user.IsAdmin() {
		role = "Администратор"
	}

	h.render(w, "profile", &ProfileData{
		PageTitle: "Профиль — TOT",
		User:      user,
		Initial:   firstInitial(user.Name, user.Email),
		Joined:    user.CreatedAt.Format("02.01.2006"),
		Completed: completed,
		Started:   len(prog),
		RoleLabel: role,
	})
}
