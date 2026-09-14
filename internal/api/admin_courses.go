package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/content"
	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxCoverBytes = 4 << 20

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// adminCourseRow is AdminCourseRow built on the AdminCourse fields.
type adminCourseRow struct {
	apigen.AdminCourse
	LessonsCount int `json:"lessons_count"`
	LabsCount    int `json:"labs_count"`
}

func (a *API) adminListCourses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := userFrom(ctx)
	rows, err := a.Modules.ListManaged(ctx, user.ID, user.IsAdmin())
	if err != nil {
		a.internalError(w, "admin: list courses", err)
		return
	}
	owners := map[int]*apigen.AuthorRef{}
	out := make([]adminCourseRow, 0, len(rows))
	for _, row := range rows {
		level := levelCoauthor
		switch {
		case user.IsAdmin():
			level = levelAdmin
		case row.OwnerID != nil && *row.OwnerID == user.ID:
			level = levelOwner
		}
		owner, err := a.authorRef(ctx, row.OwnerID, owners)
		if err != nil {
			a.internalError(w, "admin: course owner", err)
			return
		}
		out = append(out, adminCourseRow{AdminCourse: toAdminCourse(row.Module, owner, level), LessonsCount: row.Lessons, LabsCount: row.Labs})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) adminGetCourse(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	ctx := r.Context()
	lessons, err := a.Lessons.GetByModuleAll(ctx, course.module.ID)
	if err != nil {
		a.internalError(w, "admin: course lessons", err)
		return
	}
	rows := make([]apigen.AdminLessonRow, 0, len(lessons))
	for _, l := range lessons {
		questions, tasks := a.Lessons.CountsForLesson(ctx, l.ID)
		rows = append(rows, apigen.AdminLessonRow{
			ID: l.ID, Slug: l.Slug, Title: l.Title, Kind: lessonKind(l.Kind), Published: l.Published,
			VMImage: l.VMImage, QuestionsCount: questions, TasksCount: tasks,
		})
	}
	a.writeAdminCourse(w, r, http.StatusOK, course, func(c apigen.AdminCourse) any {
		return apigen.AdminCourseDetail{Course: c, Lessons: rows}
	})
}

func (a *API) adminCreateCourse(w http.ResponseWriter, r *http.Request) {
	var body apigen.AdminCourseInput
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	user := userFrom(ctx)
	m, fields, err := a.courseFromInput(ctx, body, nil)
	if err != nil {
		a.internalError(w, "admin: validate course", err)
		return
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	if m.OrderNum, err = a.Modules.NextOrder(ctx); err != nil {
		a.internalError(w, "admin: course order", err)
		return
	}
	if body.CoverURL != nil {
		m.CoverImage = *body.CoverURL
	}
	m.Published = body.Published != nil && *body.Published
	m.OwnerID = &user.ID
	id, err := a.Modules.Create(ctx, m)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeSlugTaken, "course slug is taken")
		return
	}
	if err != nil {
		a.internalError(w, "admin: create course", err)
		return
	}
	course, ok := a.managedCourse(w, r, id, needEdit)
	if !ok {
		return
	}
	a.writeAdminCourse(w, r, http.StatusCreated, course, nil)
}

func (a *API) adminUpdateCourse(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	var body apigen.AdminCourseInput
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	cur := course.module
	m, fields, err := a.courseFromInput(ctx, body, cur)
	if err != nil {
		a.internalError(w, "admin: validate course", err)
		return
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	m.ID, m.OrderNum, m.OwnerID, m.Source, m.CreatedAt = cur.ID, cur.OrderNum, cur.OwnerID, cur.Source, cur.CreatedAt
	m.CoverImage, m.Published = cur.CoverImage, cur.Published
	if body.CoverURL != nil {
		m.CoverImage = *body.CoverURL
	}
	if body.Published != nil && *body.Published != cur.Published {
		if course.level < needPublish {
			writeError(w, http.StatusForbidden, codeForbidden, "only the owner can publish the course")
			return
		}
		m.Published = *body.Published
	}
	err = a.Modules.Update(ctx, m)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeSlugTaken, "course slug is taken")
		return
	}
	if err != nil {
		a.internalError(w, "admin: update course", err)
		return
	}
	course.module = &m
	a.writeAdminCourse(w, r, http.StatusOK, course, nil)
}

func (a *API) adminDeleteCourse(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needPublish)
	if !ok {
		return
	}
	if err := a.Modules.Delete(r.Context(), course.module.ID); err != nil {
		a.internalError(w, "admin: delete course", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminSetCoursePublished(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needPublish)
	if !ok {
		return
	}
	var body apigen.Published
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := a.Modules.SetPublished(r.Context(), course.module.ID, body.Published); err != nil {
		a.internalError(w, "admin: publish course", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminMoveCourse(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needAdmin)
	if !ok {
		return
	}
	dir, ok := decodeMove(w, r)
	if !ok {
		return
	}
	if err := a.Modules.Move(r.Context(), course.module.ID, dir); err != nil {
		a.internalError(w, "admin: move course", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminUploadCourseCover(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	cover, ok := readCoverUpload(w, r)
	if !ok {
		return
	}
	if err := a.Modules.SetCover(r.Context(), course.module.ID, cover); err != nil {
		a.internalError(w, "admin: upload course cover", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminDeleteCourseCover(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	if err := a.Modules.SetCover(r.Context(), course.module.ID, ""); err != nil {
		a.internalError(w, "admin: delete course cover", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminListCourseAuthors(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	a.writeCourseAuthors(w, r, course.module.ID)
}

func (a *API) adminAddCourseAuthor(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needPublish)
	if !ok {
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	email := strings.TrimSpace(body.Email)
	u, err := a.Users.GetByEmail(ctx, email)
	if isNotFound(err) && email != strings.ToLower(email) {
		u, err = a.Users.GetByEmail(ctx, strings.ToLower(email))
	}
	if isNotFound(err) {
		writeValidation(w, map[string]string{"email": fieldNotFound})
		return
	}
	if err != nil {
		a.internalError(w, "admin: find author", err)
		return
	}
	if !u.CanAuthor() {
		writeValidation(w, map[string]string{"email": fieldNotAuthor})
		return
	}
	if owner := course.module.OwnerID; owner != nil && *owner == u.ID {
		writeError(w, http.StatusConflict, codeAlreadyAuthor, "user owns the course")
		return
	}
	created, err := a.Authors.Add(ctx, course.module.ID, u.ID, userFrom(ctx).ID)
	if err != nil {
		a.internalError(w, "admin: add author", err)
		return
	}
	if !created {
		writeError(w, http.StatusConflict, codeAlreadyAuthor, "user already co-authors the course")
		return
	}
	writeJSON(w, http.StatusCreated, toAuthorRef(*u))
}

func (a *API) adminRemoveCourseAuthor(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	uid, ok := pathID(w, r, "userId", "author not found")
	if !ok {
		return
	}
	ctx := r.Context()
	if course.level < needPublish && uid != userFrom(ctx).ID {
		writeError(w, http.StatusForbidden, codeForbidden, "only the owner can remove co-authors")
		return
	}
	removed, err := a.Authors.Remove(ctx, course.module.ID, uid)
	if err != nil {
		a.internalError(w, "admin: remove author", err)
		return
	}
	if !removed {
		writeError(w, http.StatusNotFound, codeNotFound, "author not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminSetCourseOwner(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needAdmin)
	if !ok {
		return
	}
	var body apigen.AdminSetCourseOwnerJSONRequestBody
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	u, err := a.Users.GetByID(ctx, body.UserID)
	if isNotFound(err) {
		writeValidation(w, map[string]string{"user_id": fieldNotFound})
		return
	}
	if err != nil {
		a.internalError(w, "admin: find owner", err)
		return
	}
	if !u.CanAuthor() {
		writeValidation(w, map[string]string{"user_id": fieldNotAuthor})
		return
	}
	if err := a.Authors.SetOwner(ctx, course.module.ID, u.ID, userFrom(ctx).ID); err != nil {
		a.internalError(w, "admin: set owner", err)
		return
	}
	a.writeCourseAuthors(w, r, course.module.ID)
}

func (a *API) adminPreviewContent(w http.ResponseWriter, r *http.Request) {
	var body apigen.AdminPreviewContentJSONRequestBody
	if !decodeJSON(w, r, &body) {
		return
	}
	if !body.Format.Valid() {
		writeValidation(w, map[string]string{"format": fieldInvalidValue})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"html": content.Render(string(body.Format), body.Content)})
}

// writeAdminCourse responds with the course, optionally wrapped by wrap.
func (a *API) writeAdminCourse(w http.ResponseWriter, r *http.Request, status int, course courseRef, wrap func(apigen.AdminCourse) any) {
	owner, err := a.authorRef(r.Context(), course.module.OwnerID, nil)
	if err != nil {
		a.internalError(w, "admin: course owner", err)
		return
	}
	var out any = toAdminCourse(*course.module, owner, course.level)
	if wrap != nil {
		out = wrap(out.(apigen.AdminCourse))
	}
	writeJSON(w, status, out)
}

func (a *API) writeCourseAuthors(w http.ResponseWriter, r *http.Request, moduleID int) {
	ctx := r.Context()
	m, err := a.Modules.GetByID(ctx, moduleID)
	if err != nil {
		a.internalError(w, "admin: reload course", err)
		return
	}
	owner, err := a.authorRef(ctx, m.OwnerID, nil)
	if err != nil {
		a.internalError(w, "admin: course owner", err)
		return
	}
	coauthors, err := a.Authors.List(ctx, moduleID)
	if err != nil {
		a.internalError(w, "admin: list authors", err)
		return
	}
	out := apigen.CourseAuthors{Owner: owner, Coauthors: make([]apigen.AuthorRef, 0, len(coauthors))}
	for _, u := range coauthors {
		out.Coauthors = append(out.Coauthors, toAuthorRef(u))
	}
	writeJSON(w, http.StatusOK, out)
}

// authorRef loads a user as AuthorRef; nil id or a deleted user gives nil. cache may be nil.
func (a *API) authorRef(ctx context.Context, id *int, cache map[int]*apigen.AuthorRef) (*apigen.AuthorRef, error) {
	if id == nil {
		return nil, nil
	}
	if ref, ok := cache[*id]; ok {
		return ref, nil
	}
	u, err := a.Users.GetByID(ctx, *id)
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ref := toAuthorRef(*u)
	if cache != nil {
		cache[*id] = &ref
	}
	return &ref, nil
}

func toAuthorRef(u repository.User) apigen.AuthorRef {
	return apigen.AuthorRef{ID: u.ID, Name: u.Name, Email: openapi_types.Email(u.Email)}
}

func toAdminCourse(m model.Module, owner *apigen.AuthorRef, level courseLevel) apigen.AdminCourse {
	out := apigen.AdminCourse{
		ID: m.ID, Slug: m.Slug, Title: m.Title, Description: m.Description, Track: m.Track,
		Difficulty: apigen.Difficulty(m.Difficulty), Category: m.Category, Accent: m.Accent,
		Tags: m.Tags, EstMinutes: m.EstMinutes, OrderNum: m.OrderNum, Published: m.Published,
		Source: apigen.AdminCourseSource(m.Source), Owner: owner, Access: courseAccess(level),
		HasCustomCover:  m.CoverImage != "",
		CoverPreviewURL: "/api/v1/courses/" + m.Slug + "/cover",
		CreatedAt:       m.CreatedAt,
	}
	if out.Tags == nil {
		out.Tags = []string{}
	}
	if m.Label != "" {
		code := apigen.CourseLabel(catalog.LabelCode(m))
		out.Label = &code
	}
	if m.CoverImage != "" && !strings.HasPrefix(m.CoverImage, "data:") {
		out.CoverURL = &m.CoverImage
	}
	return out
}

// courseFromInput validates a course body; cur is the course being updated or nil on create.
func (a *API) courseFromInput(ctx context.Context, in apigen.AdminCourseInput, cur *model.Module) (model.Module, map[string]string, error) {
	fields := map[string]string{}
	m := model.Module{
		Slug:        strings.TrimSpace(in.Slug),
		Title:       strings.TrimSpace(in.Title),
		Description: deref(in.Description),
		Track:       strings.TrimSpace(in.Track),
		Difficulty:  string(in.Difficulty),
		Category:    strings.TrimSpace(deref(in.Category)),
		Accent:      strings.TrimSpace(deref(in.Accent)),
		EstMinutes:  deref(in.EstMinutes),
		Tags:        []string{},
	}
	checkSlug(fields, "slug", m.Slug)
	checkText(fields, "title", m.Title, 200, true)
	checkText(fields, "description", m.Description, 5000, false)
	checkText(fields, "category", m.Category, 40, false)
	checkText(fields, "accent", m.Accent, 40, false)
	if !in.Difficulty.Valid() {
		fields["difficulty"] = fieldInvalidValue
	}
	if in.Label != nil {
		if !in.Label.Valid() {
			fields["label"] = fieldInvalidValue
		}
		m.Label = string(*in.Label)
	}
	if m.EstMinutes < 0 || m.EstMinutes > 100000 {
		fields["est_minutes"] = fieldInvalidValue
	}
	if in.Tags != nil {
		for _, t := range *in.Tags {
			if t = strings.TrimSpace(t); t != "" {
				m.Tags = append(m.Tags, t)
			}
			if utf8.RuneCountInString(t) > 40 {
				fields["tags"] = fieldTooLong
			}
		}
		if len(m.Tags) > 20 {
			fields["tags"] = fieldTooLong
		}
	}
	if in.CoverURL != nil && !isHTTPURL(*in.CoverURL) {
		fields["cover_url"] = fieldInvalidFormat
	}
	switch {
	case m.Track == "":
		fields["track"] = fieldRequired
	case cur != nil && cur.Track == m.Track:
	case m.Track == catalog.GymSpec:
		if !userFrom(ctx).IsAdmin() {
			fields["track"] = fieldInvalidValue
		}
	default:
		if _, err := a.Specs.Get(ctx, m.Track); isNotFound(err) {
			fields["track"] = fieldNotFound
		} else if err != nil {
			return m, nil, err
		}
	}
	return m, fields, nil
}

func checkSlug(fields map[string]string, name, slug string) {
	switch {
	case slug == "":
		fields[name] = fieldRequired
	case len(slug) > 100:
		fields[name] = fieldTooLong
	case !slugPattern.MatchString(slug):
		fields[name] = fieldInvalidFormat
	}
}

func checkText(fields map[string]string, name, value string, max int, required bool) {
	switch {
	case required && strings.TrimSpace(value) == "":
		fields[name] = fieldRequired
	case utf8.RuneCountInString(value) > max:
		fields[name] = fieldTooLong
	}
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && len(s) <= 2000 && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// decodeMove reads a Move body, answering 422 for an unknown direction.
func decodeMove(w http.ResponseWriter, r *http.Request) (string, bool) {
	var body apigen.Move
	if !decodeJSON(w, r, &body) {
		return "", false
	}
	if !body.Direction.Valid() {
		writeValidation(w, map[string]string{"direction": fieldInvalidValue})
		return "", false
	}
	return string(body.Direction), true
}

// readCoverUpload reads the multipart "file" part as a data URI, writing the error response itself when it returns false.
func readCoverUpload(w http.ResponseWriter, r *http.Request) (string, bool) {
	if ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || ct != "multipart/form-data" {
		writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMedia, "expected multipart/form-data body")
		return "", false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxCoverBytes+64<<10)
	mr, err := r.MultipartReader()
	if err != nil {
		writeValidation(w, map[string]string{"file": fieldRequired})
		return "", false
	}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", coverReadError(w, err)
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, maxCoverBytes+1))
		if err != nil {
			return "", coverReadError(w, err)
		}
		if len(data) > maxCoverBytes {
			writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, "cover exceeds 4 MiB")
			return "", false
		}
		mimeType, ok := coverMIME(data)
		if !ok {
			writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMedia, "cover must be PNG, JPEG, WebP or SVG")
			return "", false
		}
		return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data), true
	}
	writeValidation(w, map[string]string{"file": fieldRequired})
	return "", false
}

func coverReadError(w http.ResponseWriter, err error) bool {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, "cover exceeds 4 MiB")
		return false
	}
	writeValidation(w, map[string]string{"file": fieldInvalidFormat})
	return false
}

// coverMIME detects a supported cover image type from its bytes.
func coverMIME(data []byte) (string, bool) {
	switch ct := http.DetectContentType(data); ct {
	case "image/png", "image/jpeg", "image/webp":
		return ct, true
	}
	head := bytes.ToLower(data[:min(len(data), 1024)])
	if bytes.Contains(head, []byte("<svg")) {
		return "image/svg+xml", true
	}
	return "", false
}
