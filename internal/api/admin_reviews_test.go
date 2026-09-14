package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
)

func TestCourseReviewFlow(t *testing.T) {
	h, c := authorsFixture(t)
	req := func(token, method, path, body string) (int, string) {
		w := do(h, method, path, body, withCookie(token))
		return w.Code, w.Body.String()
	}
	course := func(token string) apigen.AdminCourse {
		return decode[apigen.AdminCourseDetail](t, do(h, http.MethodGet, "/admin/courses/12", "", withCookie(token))).Course
	}

	if code, _ := req(coToken, http.MethodPost, "/admin/courses/12/review", ""); code != http.StatusForbidden {
		t.Errorf("co-author requests review: %d", code)
	}
	code, body := req(ownerToken, http.MethodPost, "/admin/courses/12/review", `{"note":" Ready "}`)
	if code != http.StatusCreated {
		t.Fatalf("request review: %d %s", code, body)
	}
	if code, body := req(ownerToken, http.MethodPost, "/admin/courses/12/review", ""); code != http.StatusConflict || !strings.Contains(body, codeReviewPending) {
		t.Errorf("second request: %d %s", code, body)
	}
	review := course(coToken).Review
	if review == nil || review.Status != apigen.Pending || review.Note != "Ready" || review.Course.Slug != "k8s-draft" || review.Lesson != nil || review.RequestedBy == nil || review.RequestedBy.ID != 3 {
		t.Fatalf("pending review = %+v", review)
	}
	path := "/admin/reviews/" + itoa(review.ID)

	if code, _ := req(ownerToken, http.MethodPost, path+"/approve", ""); code != http.StatusForbidden {
		t.Errorf("owner approves own review: %d", code)
	}
	if list := decode[[]apigen.ReviewRequest](t, do(h, http.MethodGet, "/admin/reviews?status=pending", "", withCookie(otherToken))); len(list) != 0 {
		t.Errorf("unrelated author sees reviews: %+v", list)
	}
	if list := decode[[]apigen.ReviewRequest](t, do(h, http.MethodGet, "/admin/reviews?status=pending", "", withCookie(coToken))); len(list) != 1 {
		t.Errorf("co-author reviews: %+v", list)
	}
	if code, body := req(adminToken, http.MethodPost, path+"/reject", `{}`); code != http.StatusUnprocessableEntity || !strings.Contains(body, `"note":"required"`) {
		t.Errorf("reject without note: %d %s", code, body)
	}
	code, body = req(adminToken, http.MethodPost, path+"/reject", `{"note":"Add a quiz"}`)
	if code != http.StatusOK || !strings.Contains(body, `"status":"rejected"`) || !strings.Contains(body, "Add a quiz") || c.modules[2].Published {
		t.Fatalf("reject: %d %s", code, body)
	}
	if got := course(ownerToken); got.Review == nil || got.Review.Status != apigen.Rejected || got.Review.DecidedBy == nil {
		t.Errorf("rejected review on course = %+v", got.Review)
	}
	if code, body := req(adminToken, http.MethodPost, path+"/approve", ""); code != http.StatusConflict || !strings.Contains(body, codeReviewClosed) {
		t.Errorf("approve closed review: %d %s", code, body)
	}

	code, body = req(ownerToken, http.MethodPost, "/admin/courses/12/review", "")
	if code != http.StatusCreated {
		t.Fatalf("resubmit: %d %s", code, body)
	}
	resubmitted := course(ownerToken).Review
	if code, body := req(adminToken, http.MethodPost, "/admin/reviews/"+itoa(resubmitted.ID)+"/approve", `{"note":"Good"}`); code != http.StatusOK || !strings.Contains(body, `"status":"approved"`) {
		t.Fatalf("approve: %d %s", code, body)
	}
	if got := course(ownerToken); !got.Published || got.Review != nil {
		t.Errorf("approved course = published %v review %+v", got.Published, got.Review)
	}
	if code, body := req(ownerToken, http.MethodPost, "/admin/courses/12/review", ""); code != http.StatusConflict || !strings.Contains(body, codeReviewNotNeeded) {
		t.Errorf("review of a published course: %d %s", code, body)
	}
	if w := do(h, http.MethodGet, "/courses/k8s-draft", "", withCookie(studentToken)); w.Code != http.StatusOK {
		t.Errorf("student sees approved course: %d", w.Code)
	}
}

func TestLessonReviewAndAdminAuthor(t *testing.T) {
	h, c := authorsFixture(t)
	admin := 2
	c.modules[0].OwnerID = &admin
	c.lessons = append(c.lessons, model.Lesson{ID: 140, ModuleID: 10, Slug: "new", Title: "New", Kind: "theory", OrderNum: 5})
	req := func(token, method, path, body string) (int, string) {
		w := do(h, method, path, body, withCookie(token))
		return w.Code, w.Body.String()
	}

	if code, body := req(adminToken, http.MethodPut, "/admin/lessons/140/published", `{"published":true}`); code != http.StatusForbidden || !strings.Contains(body, codeReviewRequired) {
		t.Errorf("admin author publishes lesson of a published course: %d %s", code, body)
	}
	if code, _ := req(adminToken, http.MethodPost, "/admin/lessons/100/review", ""); code != http.StatusConflict {
		t.Errorf("review of a published lesson: %d", code)
	}
	code, body := req(adminToken, http.MethodPost, "/admin/lessons/140/review", "")
	if code != http.StatusCreated || !strings.Contains(body, `"lesson":{"id":140`) {
		t.Fatalf("admin requests lesson review: %d %s", code, body)
	}
	rows := decode[apigen.AdminCourseDetail](t, do(h, http.MethodGet, "/admin/courses/10", "", withCookie(adminToken))).Lessons
	var status *apigen.ReviewStatus
	for _, row := range rows {
		if row.ID == 140 {
			status = row.ReviewStatus
		}
	}
	if status == nil || *status != apigen.Pending {
		t.Errorf("lesson row review status = %v", status)
	}
	detail := decode[apigen.AdminLessonDetail](t, do(h, http.MethodGet, "/admin/lessons/140", "", withCookie(adminToken)))
	if detail.Review == nil {
		t.Fatalf("lesson detail review missing")
	}
	if code, body := req(adminToken, http.MethodPost, "/admin/reviews/"+itoa(detail.Review.ID)+"/approve", ""); code != http.StatusOK {
		t.Fatalf("admin approves own review: %d %s", code, body)
	}
	if l := (fakeLessons{c}).lesson(140); !l.Published {
		t.Errorf("approved lesson not published")
	}

	// lessons of an unpublished course are published directly by the owner
	if code, _ := req(ownerToken, http.MethodPut, "/admin/lessons/120/published", `{"published":false}`); code != http.StatusNoContent {
		t.Errorf("owner unpublishes lesson: %d", code)
	}
	if code, _ := req(ownerToken, http.MethodPut, "/admin/lessons/120/published", `{"published":true}`); code != http.StatusNoContent {
		t.Errorf("owner publishes lesson of a draft course: %d", code)
	}
	if code, _ := req(coToken, http.MethodPut, "/admin/lessons/120/published", `{"published":false}`); code != http.StatusForbidden {
		t.Errorf("co-author unpublishes lesson: %d", code)
	}
	if code, body := req(ownerToken, http.MethodPost, "/admin/lessons/120/review", ""); code != http.StatusConflict || !strings.Contains(body, codeReviewNotNeeded) {
		t.Errorf("lesson review in a draft course: %d %s", code, body)
	}

	code, body = req(ownerToken, http.MethodPost, "/admin/courses/12/review", "")
	if code != http.StatusCreated {
		t.Fatalf("course review: %d %s", code, body)
	}
	reviewID := c.reviews[len(c.reviews)-1].ID
	if code, _ := req(coToken, http.MethodDelete, "/admin/reviews/"+itoa(reviewID), ""); code != http.StatusForbidden {
		t.Errorf("co-author cancels: %d", code)
	}
	if code, _ := req(otherToken, http.MethodDelete, "/admin/reviews/"+itoa(reviewID), ""); code != http.StatusNotFound {
		t.Errorf("unrelated author cancels: %d", code)
	}
	if code, _ := req(ownerToken, http.MethodDelete, "/admin/reviews/"+itoa(reviewID), ""); code != http.StatusNoContent {
		t.Errorf("owner cancels: %d", code)
	}
	if code, _ := req(ownerToken, http.MethodDelete, "/admin/reviews/"+itoa(reviewID), ""); code != http.StatusConflict {
		t.Errorf("cancel twice: %d", code)
	}
	if code, _ := req(adminToken, http.MethodGet, "/admin/reviews?status=done", ""); code != http.StatusUnprocessableEntity {
		t.Errorf("bad status filter: %d", code)
	}
}
