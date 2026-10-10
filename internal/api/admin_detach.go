package api

import (
	"context"
	"net/http"

	"github.com/backendraz/golearn/internal/content"
	"github.com/backendraz/golearn/internal/model"
)

// seedDetachStore moves content ownership from the seeder to the database.
type seedDetachStore interface {
	MarkdownLessons(ctx context.Context) ([]model.Lesson, error)
	SetLessonContent(ctx context.Context, id int, format, content string) error
	Detach(ctx context.Context) (lessons, modules int64, err error)
}

type detachResponse struct {
	Converted       int   `json:"converted"`
	LessonsDetached int64 `json:"lessons_detached"`
	ModulesDetached int64 `json:"modules_detached"`
}

// detachFromSeed makes the database the source of truth for every course.
//
// Two things, in this order and for the same reason. Every Markdown lesson is
// rendered to the HTML the site already shows, so one editor opens them all;
// then edited_at is stamped on everything, so the seeder stops rewriting them
// from cmd/seed/content on each deploy. Doing only the first would be undone
// by the next release.
//
// One way. Afterwards the JSON files no longer reach the site and courses are
// written in the studio, which is the point: content stops living in two
// places that can disagree.
//
// Admin only, and idempotent — a second run converts nothing and stamps
// nothing, because there is nothing left in either state.
func (a *API) detachFromSeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	md, err := a.SeedDetach.MarkdownLessons(ctx)
	if err != nil {
		a.internalError(w, "detach: list markdown lessons", err)
		return
	}
	for _, l := range md {
		// The same renderer the lesson page uses, so what is stored is what was
		// already on the screen.
		if err := a.SeedDetach.SetLessonContent(ctx, l.ID, "html", content.Render("md", l.Content)); err != nil {
			a.internalError(w, "detach: convert lesson", err)
			return
		}
	}

	lessons, modules, err := a.SeedDetach.Detach(ctx)
	if err != nil {
		a.internalError(w, "detach: stamp edited_at", err)
		return
	}
	a.log.Info("content detached from the seeder",
		"converted", len(md), "lessons", lessons, "modules", modules,
		"by", userFrom(ctx).ID)

	writeJSON(w, http.StatusOK, detachResponse{
		Converted: len(md), LessonsDetached: lessons, ModulesDetached: modules,
	})
}
