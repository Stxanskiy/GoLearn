package api

import (
	"context"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/model"
)

// prereqState resolves a course's prerequisites against one student's progress.
//
// locked is true when a *required* prerequisite is unfinished. The course page
// still renders in that state - a student has to be able to see what is missing
// - so the block is enforced where the content is, on the lesson endpoints.
//
// A prerequisite that names a course which no longer exists is dropped rather
// than treated as unmet: a deleted or renamed course would otherwise lock
// everything that pointed at it, with nothing in the interface to resolve.
func (a *API) prereqState(ctx context.Context, m model.Module, up catalog.UserProgress) ([]apigen.Prerequisite, bool, error) {
	if len(m.Prerequisites) == 0 {
		return []apigen.Prerequisite{}, false, nil
	}
	out := make([]apigen.Prerequisite, 0, len(m.Prerequisites))
	locked := false
	for _, p := range m.Prerequisites {
		if p.Slug == m.Slug {
			// A course cannot gate itself; such a row would lock it forever.
			continue
		}
		other, err := a.Modules.GetBySlug(ctx, p.Slug)
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return nil, false, err
		}
		lessons, err := a.Lessons.GetByModule(ctx, other.ID)
		if err != nil {
			return nil, false, err
		}
		c := catalog.BuildCourse(*other, lessons, up)
		// An empty course can never be completed, so requiring one would be a
		// permanent lock. Treat "nothing to do" as done.
		done := len(lessons) == 0 || c.Status == catalog.StatusCompleted
		pct := c.Pct
		kind := apigen.PrerequisiteKind(p.Kind)
		if !p.Required() {
			kind = apigen.Recommended
		} else {
			kind = apigen.Required
			if !done {
				locked = true
			}
		}
		out = append(out, apigen.Prerequisite{
			Slug:        other.Slug,
			Title:       other.Title,
			Kind:        kind,
			Completed:   done,
			ProgressPct: &pct,
		})
	}
	return out, locked, nil
}

// courseLocked reports whether this student still has a required prerequisite
// to finish. Used by the lesson endpoints, which must refuse the content.
func (a *API) courseLocked(ctx context.Context, m model.Module, userID int) (bool, error) {
	if len(m.Prerequisites) == 0 {
		return false, nil
	}
	up, err := a.userProgress(ctx, userID)
	if err != nil {
		return false, err
	}
	_, locked, err := a.prereqState(ctx, m, up)
	return locked, err
}
