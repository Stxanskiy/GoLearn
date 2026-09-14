package api

import (
	"net/http"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
)

const activityDays = 365

func (a *API) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := userFrom(ctx)
	courses, err := a.publishedCourses(ctx, user.ID)
	if err != nil {
		a.internalError(w, "profile: courses", err)
		return
	}
	out := apigen.Profile{User: toMe(user)}
	for _, c := range courses {
		for _, l := range c.Lessons {
			switch l.Status {
			case catalog.StatusCompleted:
				out.LessonsCompleted++
				out.LessonsStarted++
			case catalog.StatusInProgress:
				out.LessonsStarted++
			}
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) getDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := userFrom(ctx).ID
	ov, err := a.Progress.Overview(ctx, uid)
	if err != nil {
		a.internalError(w, "dashboard: overview", err)
		return
	}
	courses, err := a.publishedCourses(ctx, uid)
	if err != nil {
		a.internalError(w, "dashboard: courses", err)
		return
	}
	sims, err := a.Sims.ListPublished(ctx)
	if err != nil {
		a.internalError(w, "dashboard: simulators", err)
		return
	}
	cont, err := a.Progress.LatestInProgress(ctx, uid)
	if err != nil {
		a.internalError(w, "dashboard: continue", err)
		return
	}

	out := apigen.Dashboard{
		Overview: apigen.DashboardOverview{
			Streak:          ov.Streak,
			ActiveDays:      ov.ActiveDays,
			TasksSolved:     ov.TasksSolved,
			LabsDone:        ov.LabsDone,
			LabsTotal:       ov.LabsTotal,
			LessonsRead:     ov.ArticlesRead,
			LessonsTotal:    ov.ArticlesTotal,
			TestsPassed:     ov.TestsPassed,
			SimulatorsTotal: len(sims),
		},
		Activity: recentActivity(ov.Activity, time.Now()),
	}
	for _, c := range courses {
		if c.Spec != catalog.GymSpec {
			continue
		}
		out.Overview.TrainersTotal++
		if c.Status == catalog.StatusCompleted {
			out.Overview.TrainersDone++
		}
	}
	if cont != nil {
		out.Continue = &apigen.ContinueLesson{
			Course:   apigen.LinkRef{Slug: cont.ModuleSlug, Title: cont.ModuleTitle},
			Lesson:   apigen.LinkRef{Slug: cont.LessonSlug, Title: cont.LessonTitle},
			LessonID: cont.LessonID,
			Kind:     lessonKind(cont.LessonKind),
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// recentActivity keeps the last activityDays days (today included).
func recentActivity(all map[string]int, now time.Time) map[string]int {
	from := now.AddDate(0, 0, -(activityDays - 1)).Format(time.DateOnly)
	out := make(map[string]int)
	for day, n := range all {
		if day >= from {
			out[day] = n
		}
	}
	return out
}
