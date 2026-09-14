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
	if review == nil || review.Status != apigen.Pending || review.Note != "Ready" || review.Course.Slug != "k8s-draft" || review.RequestedBy == nil || review.RequestedBy.ID != 3 {
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

func TestPublishedCourseDraftFlow(t *testing.T) {
	h, c := authorsFixture(t)
	owner := 3
	c.modules[0].OwnerID = &owner // linux: published, owned by the author
	c.coauthors[10] = []int{4}
	c.tasks[102] = []model.Task{{ID: 900, LessonID: 102, Title: "Run", Kind: "shell", Format: "md", Difficulty: "easy"}}
	c.passed[1] = map[int]bool{900: true}
	req := func(token, method, path, body string) (int, string) {
		w := do(h, method, path, body, withCookie(token))
		return w.Code, w.Body.String()
	}
	lesson := `{"slug":"extra","title":"Extra","kind":"theory","format":"md","difficulty":"beginner","published":true}`

	for _, tt := range []struct{ method, path, body string }{
		{http.MethodPut, "/admin/courses/10", `{"slug":"linux","title":"Linux 2","track":"devops","difficulty":"beginner"}`},
		{http.MethodPost, "/admin/courses/10/lessons", lesson},
		{http.MethodPost, "/admin/tasks/900", ``},
		{http.MethodDelete, "/admin/lessons/100", ``},
		{http.MethodPut, "/admin/lessons/103/published", `{"published":true}`},
	} {
		if code, body := req(adminToken, tt.method, tt.path, tt.body); tt.path != "/admin/tasks/900" && (code != http.StatusConflict || !strings.Contains(body, codeDraftRequired)) {
			t.Errorf("admin edits published course %s %s: %d %s", tt.method, tt.path, code, body)
		}
	}
	if code, _ := req(ownerToken, http.MethodPut, "/admin/lessons/101/published", `{"published":false}`); code != http.StatusNoContent {
		t.Errorf("owner hides a lesson of a published course: %d", code)
	}
	c.lessons[1].Published = true
	if code, body := req(ownerToken, http.MethodPost, "/admin/courses/10/review", ""); code != http.StatusConflict || !strings.Contains(body, codeReviewNotNeeded) {
		t.Errorf("review without draft: %d %s", code, body)
	}
	if code, _ := req(ownerToken, http.MethodPost, "/admin/courses/12/draft", ""); code != http.StatusConflict {
		t.Errorf("draft of an unpublished course: %d", code)
	}

	code, body := req(coToken, http.MethodPost, "/admin/courses/10/draft", "")
	if code != http.StatusCreated {
		t.Fatalf("open draft: %d %s", code, body)
	}
	draft := decode[apigen.AdminCourse](t, do(h, http.MethodPost, "/admin/courses/10/draft", "", withCookie(coToken)))
	if draft.DraftOf == nil || *draft.DraftOf != 10 || draft.Slug != "linux" || draft.PreviewSlug == "linux" || draft.Published {
		t.Fatalf("draft = %+v", draft)
	}
	if live := decode[apigen.AdminCourseDetail](t, do(h, http.MethodGet, "/admin/courses/10", "", withCookie(coToken))).Course; live.DraftID == nil || *live.DraftID != draft.ID {
		t.Errorf("live course draft id = %v", live.DraftID)
	}
	for _, row := range decode[[]apigen.AdminCourseRow](t, do(h, http.MethodGet, "/admin/courses", "", withCookie(coToken))) {
		if row.ID == draft.ID {
			t.Errorf("drafts must not be listed: %+v", row)
		}
	}
	dpath := "/admin/courses/" + itoa(draft.ID)
	if code, body := req(coToken, http.MethodPut, dpath, `{"slug":"linux-2","title":"Linux 2","track":"devops","difficulty":"beginner"}`); code != http.StatusUnprocessableEntity || !strings.Contains(body, `"slug":"invalid_value"`) {
		t.Errorf("rename slug in draft: %d %s", code, body)
	}
	if code, body := req(coToken, http.MethodPut, dpath, `{"slug":"linux","title":"Linux 2","track":"devops","difficulty":"beginner"}`); code != http.StatusOK {
		t.Fatalf("edit draft: %d %s", code, body)
	}
	if code, body := req(ownerToken, http.MethodPost, dpath+"/lessons", lesson); code != http.StatusCreated {
		t.Fatalf("owner adds a published lesson to the draft: %d %s", code, body)
	}
	if code, _ := req(ownerToken, http.MethodPut, dpath+"/published", `{"published":true}`); code != http.StatusForbidden {
		t.Errorf("publish draft directly: %d", code)
	}
	if code, body := req(ownerToken, http.MethodPut, "/admin/courses/10/published", `{"published":false}`); code != http.StatusConflict || !strings.Contains(body, codeDraftExists) {
		t.Errorf("unpublish course with draft: %d %s", code, body)
	}
	if w := do(h, http.MethodGet, "/courses/"+draft.PreviewSlug+"/lessons/extra", "", withCookie(coToken)); w.Code != http.StatusOK {
		t.Errorf("co-author previews draft lesson: %d", w.Code)
	}
	if w := do(h, http.MethodGet, "/courses/"+draft.PreviewSlug, "", withCookie(studentToken)); w.Code != http.StatusNotFound {
		t.Errorf("student opens draft: %d", w.Code)
	}
	if live := do(h, http.MethodGet, "/courses/linux", "", withCookie(studentToken)); strings.Contains(live.Body.String(), "Linux 2") || strings.Contains(live.Body.String(), "extra") {
		t.Errorf("student sees unapproved changes: %s", live.Body)
	}

	changes := decode[apigen.DraftChanges](t, do(h, http.MethodGet, "/admin/courses/10/draft/changes", "", withCookie(adminToken)))
	if strings.Join(changes.CourseFields, ",") != "title" || len(changes.Lessons) != 1 || changes.Lessons[0].Change != "added" {
		t.Errorf("changes = %+v", changes)
	}
	if code, body = req(coToken, http.MethodPost, dpath+"/review", ""); code != http.StatusForbidden {
		t.Errorf("co-author requests review: %d %s", code, body)
	}
	code, body = req(ownerToken, http.MethodPost, dpath+"/review", `{"note":"new lesson"}`)
	if code != http.StatusCreated || !strings.Contains(body, `"kind":"changes"`) || !strings.Contains(body, `"id":10,"slug":"linux"`) {
		t.Fatalf("request changes review: %d %s", code, body)
	}
	if code, body := req(coToken, http.MethodPut, dpath, `{"slug":"linux","title":"Linux 3","track":"devops","difficulty":"beginner"}`); code != http.StatusConflict || !strings.Contains(body, codeReviewPending) {
		t.Errorf("edit draft under review: %d %s", code, body)
	}
	if code, body := req(coToken, http.MethodPost, "/admin/import", `{"slug":"linux","title":"X","track":"devops","difficulty":"beginner","lessons":[]}`); code != http.StatusConflict || !strings.Contains(body, codeReviewPending) {
		t.Errorf("import into course under review: %d %s", code, body)
	}

	review := c.reviews[len(c.reviews)-1]
	if code, body := req(adminToken, http.MethodPost, "/admin/reviews/"+itoa(review.ID)+"/approve", ""); code != http.StatusOK {
		t.Fatalf("approve changes: %d %s", code, body)
	}
	live := fakeModules{c}.module(10)
	if live.Title != "Linux 2" || !live.Published {
		t.Errorf("live course after approval = %+v", live)
	}
	if _, err := (fakeModules{c}).DraftFor(t.Context(), 10); err == nil {
		t.Errorf("draft kept after approval")
	}
	if w := do(h, http.MethodGet, "/courses/linux/lessons/extra", "", withCookie(studentToken)); w.Code != http.StatusOK {
		t.Errorf("student opens approved lesson: %d", w.Code)
	}
	if tasks := c.tasks[102]; len(tasks) != 1 || tasks[0].ID != 900 {
		t.Errorf("live task id changed: %+v", tasks)
	}

	// a second draft can be discarded, which cancels its review
	req(ownerToken, http.MethodPost, "/admin/courses/10/draft", "")
	req(ownerToken, http.MethodPost, "/admin/courses/10/review", "")
	if code, _ := req(coToken, http.MethodDelete, "/admin/courses/10/draft", ""); code != http.StatusForbidden {
		t.Errorf("co-author discards: %d", code)
	}
	if code, _ := req(ownerToken, http.MethodDelete, "/admin/courses/10/draft", ""); code != http.StatusNoContent {
		t.Errorf("discard: %d", code)
	}
	if last := c.reviews[len(c.reviews)-1]; last.Status != "cancelled" {
		t.Errorf("review after discard = %s", last.Status)
	}
	if code, _ := req(ownerToken, http.MethodDelete, "/admin/courses/10/draft", ""); code != http.StatusNotFound {
		t.Errorf("discard missing draft: %d", code)
	}
}

func TestAdminAuthorApprovesOwnReview(t *testing.T) {
	h, c := authorsFixture(t)
	admin := 2
	c.modules[2].OwnerID = &admin
	if w := do(h, http.MethodPut, "/admin/courses/12/published", `{"published":true}`, withCookie(adminToken)); w.Code != http.StatusForbidden || errorCode(t, w) != codeReviewRequired {
		t.Errorf("admin author publishes directly: %d %s", w.Code, w.Body)
	}
	w := do(h, http.MethodPost, "/admin/courses/12/review", "", withCookie(adminToken))
	if w.Code != http.StatusCreated {
		t.Fatalf("request: %d %s", w.Code, w.Body)
	}
	id := decode[apigen.ReviewRequest](t, w).ID
	if w := do(h, http.MethodPost, "/admin/reviews/"+itoa(id)+"/approve", "", withCookie(adminToken)); w.Code != http.StatusOK || !c.modules[2].Published {
		t.Errorf("approve own review: %d %s", w.Code, w.Body)
	}
	if code := do(h, http.MethodGet, "/admin/reviews?status=done", "", withCookie(adminToken)).Code; code != http.StatusUnprocessableEntity {
		t.Errorf("bad status filter: %d", code)
	}
}
