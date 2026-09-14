package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/repository"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxReviewNote = 2000

func (a *API) adminRequestCourseReview(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needOwner)
	if !ok {
		return
	}
	note, ok := decodeReviewNote(w, r, false)
	if !ok {
		return
	}
	kind := repository.ReviewPublish
	if course.live.Published {
		kind = repository.ReviewChanges
		if _, err := a.Modules.DraftFor(r.Context(), course.live.ID); isNotFound(err) {
			writeError(w, http.StatusConflict, codeReviewNotNeeded, "published course has no draft to review")
			return
		} else if err != nil {
			a.internalError(w, "admin: find draft", err)
			return
		}
	}
	ctx := r.Context()
	id, err := a.Reviews.Create(ctx, course.live.ID, kind, userFrom(ctx).ID, note)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeReviewPending, "a review is already pending")
		return
	}
	if err != nil {
		a.internalError(w, "admin: create review", err)
		return
	}
	a.writeReview(w, r, http.StatusCreated, id)
}

func (a *API) adminListReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	user := userFrom(ctx)
	f := repository.ReviewFilter{Status: query.Get("status")}
	if f.Status != "" && !apigen.ReviewStatus(f.Status).Valid() {
		writeValidation(w, map[string]string{"status": fieldInvalidValue})
		return
	}
	if v := query.Get("course_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeValidation(w, map[string]string{"course_id": fieldInvalidFormat})
			return
		}
		f.ModuleID = id
	}
	if !user.IsAdmin() {
		f.EditorID = user.ID
	}
	reviews, err := a.Reviews.List(ctx, f)
	if err != nil {
		a.internalError(w, "admin: list reviews", err)
		return
	}
	out := make([]apigen.ReviewRequest, 0, len(reviews))
	for _, v := range reviews {
		out = append(out, toReviewRequest(v))
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) adminCancelReview(w http.ResponseWriter, r *http.Request) {
	v, ok := a.reviewByID(w, r)
	if !ok {
		return
	}
	if _, ok := a.managedCourse(w, r, v.ModuleID, needOwner); !ok {
		return
	}
	if err := a.Reviews.Cancel(r.Context(), v.ID); err != nil {
		a.reviewError(w, "admin: cancel review", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminApproveReview(w http.ResponseWriter, r *http.Request) {
	a.decideReview(w, r, true)
}

func (a *API) adminRejectReview(w http.ResponseWriter, r *http.Request) {
	a.decideReview(w, r, false)
}

func (a *API) decideReview(w http.ResponseWriter, r *http.Request, approve bool) {
	v, ok := a.reviewByID(w, r)
	if !ok {
		return
	}
	note, ok := decodeReviewNote(w, r, !approve)
	if !ok {
		return
	}
	ctx := r.Context()
	if err := a.Reviews.Decide(ctx, v.ID, approve, userFrom(ctx).ID, note); errors.Is(err, repository.ErrNoDraft) {
		writeError(w, http.StatusConflict, codeReviewClosed, "the draft no longer exists")
		return
	} else if err != nil {
		a.reviewError(w, "admin: decide review", err)
		return
	}
	a.writeReview(w, r, http.StatusOK, v.ID)
}

// reviewByID loads the reviewId request, writing 404/500 itself when it returns false.
func (a *API) reviewByID(w http.ResponseWriter, r *http.Request) (*repository.Review, bool) {
	id, ok := pathID(w, r, "reviewId", "review not found")
	if !ok {
		return nil, false
	}
	v, err := a.Reviews.Get(r.Context(), id)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "review not found")
		return nil, false
	}
	if err != nil {
		a.internalError(w, "admin: load review", err)
		return nil, false
	}
	return v, true
}

func (a *API) writeReview(w http.ResponseWriter, r *http.Request, status, id int) {
	v, err := a.Reviews.Get(r.Context(), id)
	if err != nil {
		a.internalError(w, "admin: reload review", err)
		return
	}
	writeJSON(w, status, toReviewRequest(*v))
}

func (a *API) reviewError(w http.ResponseWriter, op string, err error) {
	if errors.Is(err, repository.ErrReviewClosed) {
		writeError(w, http.StatusConflict, codeReviewClosed, "review is not pending")
		return
	}
	a.internalError(w, op, err)
}

// openReview returns the latest review of a course while it needs the author's or moderator's attention.
func openReview(reviews map[int]repository.Review, moduleID int) *apigen.ReviewRequest {
	v, ok := reviews[moduleID]
	if !ok || (v.Status != repository.ReviewPending && v.Status != repository.ReviewRejected) {
		return nil
	}
	out := toReviewRequest(v)
	return &out
}

// decodeReviewNote reads an optional {"note"} body.
func decodeReviewNote(w http.ResponseWriter, r *http.Request, required bool) (string, bool) {
	var body apigen.ReviewNote
	if !decodeOptionalJSON(w, r, &body) {
		return "", false
	}
	note := strings.TrimSpace(deref(body.Note))
	fields := map[string]string{}
	checkText(fields, "note", note, maxReviewNote, required)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return "", false
	}
	return note, true
}

func toReviewRequest(v repository.Review) apigen.ReviewRequest {
	out := apigen.ReviewRequest{
		ID: v.ID, Kind: apigen.ReviewRequestKind(v.Kind), Status: apigen.ReviewStatus(v.Status),
		Course: apigen.ContentRef{ID: v.ModuleID, Slug: v.CourseSlug, Title: v.CourseTitle},
		Note:   v.Note, DecisionNote: v.DecisionNote, CreatedAt: v.CreatedAt, DecidedAt: v.DecidedAt,
	}
	if v.RequestedBy != nil {
		out.RequestedBy = &apigen.AuthorRef{ID: *v.RequestedBy, Name: v.RequesterName, Email: openapi_types.Email(v.RequesterEmail)}
	}
	if v.DecidedBy != nil {
		out.DecidedBy = &apigen.AuthorRef{ID: *v.DecidedBy, Name: v.DeciderName, Email: openapi_types.Email(v.DeciderEmail)}
	}
	return out
}
