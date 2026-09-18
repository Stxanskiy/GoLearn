package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/repository"
)

func imagesFixture(t *testing.T) (http.Handler, *fakeContent, *fakeImages) {
	t.Helper()
	users, c := storefront()
	for i, u := range []*repository.User{
		{ID: 3, Email: "owner@example.com", Name: "Owner", Role: "author"},
		{ID: 4, Email: "co@example.com", Name: "Co", Role: "author"},
		{ID: 5, Email: "other@example.com", Name: "Other", Role: "author"},
	} {
		users.byEmail[u.Email] = u
		users.sessions[[]string{ownerToken, coToken, otherToken}[i]] = u.ID
	}
	owner := 3
	c.modules[2].OwnerID = &owner
	c.coauthors[12] = []int{4}
	c.images = newFakeImages()
	return newTestAPIWith(t, users, c), c, c.images
}

func uploadTo(h http.Handler, method, path, token, field string, data []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile(field, "image.bin")
	_, _ = fw.Write(data)
	_ = mw.Close()
	return do(h, method, path, "", withCookie(token), withBody(buf.Bytes(), mw.FormDataContentType()))
}

func testPNG() []byte { return append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...) }

func TestCourseIconUpload(t *testing.T) {
	h, c, images := imagesFixture(t)

	w := uploadTo(h, http.MethodPut, "/admin/courses/12/icon", coToken, "file", testPNG())
	if w.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}
	first := decode[apigen.UploadedImage](t, w).URL
	if c.modules[2].IconURL != first {
		t.Fatalf("stored icon = %q, want %q", c.modules[2].IconURL, first)
	}
	if _, ok := images.objects[first]; !ok {
		t.Fatal("image was not put into storage")
	}

	second := decode[apigen.UploadedImage](t, uploadTo(h, http.MethodPut, "/admin/courses/12/icon", coToken, "file", testPNG())).URL
	if second == first {
		t.Fatal("replacing an icon must create a new object")
	}
	if _, ok := images.objects[first]; ok {
		t.Fatal("replaced object was not deleted")
	}

	if w := do(h, http.MethodDelete, "/admin/courses/12/icon", "", withCookie(coToken)); w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	if c.modules[2].IconURL != "" || len(images.objects) != 0 {
		t.Fatalf("after delete: icon %q, objects %d", c.modules[2].IconURL, len(images.objects))
	}
}

func TestCourseIconUploadRejections(t *testing.T) {
	h, _, _ := imagesFixture(t)
	cases := []struct {
		name, field string
		data        []byte
		want        int
	}{
		{"not an image", "file", []byte("this is plain text, not an image at all"), http.StatusUnsupportedMediaType},
		{"missing part", "other", testPNG(), http.StatusUnprocessableEntity},
		{"too large", "file", append(testPNG(), make([]byte, maxImageBytes)...), http.StatusRequestEntityTooLarge},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if w := uploadTo(h, http.MethodPut, "/admin/courses/12/icon", coToken, tt.field, tt.data); w.Code != tt.want {
				t.Fatalf("%d, want %d (%s)", w.Code, tt.want, w.Body)
			}
		})
	}
	if w := do(h, http.MethodPut, "/admin/courses/12/icon", `{}`, withCookie(coToken)); w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("json body: %d", w.Code)
	}
	if w := uploadTo(h, http.MethodPut, "/admin/courses/12/icon", otherToken, "file", testPNG()); w.Code != http.StatusNotFound {
		t.Errorf("unrelated author: %d", w.Code)
	}
}

func TestUploadWithoutStorage(t *testing.T) {
	h, _ := authorsFixture(t) // no image store configured
	w := uploadTo(h, http.MethodPut, "/admin/courses/12/icon", coToken, "file", testPNG())
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), codeStorageDisabled) {
		t.Fatalf("icon upload: %d %s", w.Code, w.Body)
	}
	if w := uploadTo(h, http.MethodPost, "/admin/uploads/images", coToken, "file", testPNG()); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("content upload: %d %s", w.Code, w.Body)
	}
	// Covers still work: they fall back to an inline data URI.
	if w := uploadTo(h, http.MethodPut, "/admin/courses/12/cover", coToken, "file", testPNG()); w.Code != http.StatusNoContent {
		t.Fatalf("cover upload: %d %s", w.Code, w.Body)
	}
}

func TestCoverGoesToStorage(t *testing.T) {
	h, c, images := imagesFixture(t)
	w := uploadTo(h, http.MethodPut, "/admin/courses/12/cover", coToken, "file", testPNG())
	if w.Code != http.StatusNoContent {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}
	if !strings.HasPrefix(c.modules[2].CoverImage, "https://cdn.test/covers/") {
		t.Fatalf("stored cover = %q", c.modules[2].CoverImage)
	}
	if _, ok := images.objects[c.modules[2].CoverImage]; !ok {
		t.Fatal("cover was not put into storage")
	}
}

func TestSpecializationIcon(t *testing.T) {
	h, c, images := imagesFixture(t)
	slug := c.specs[0].Slug

	w := uploadTo(h, http.MethodPut, "/admin/specializations/"+slug+"/icon", adminToken, "file", testPNG())
	if w.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}
	url := decode[apigen.UploadedImage](t, w).URL
	if c.specs[0].IconURL != url {
		t.Fatalf("stored icon = %q, want %q", c.specs[0].IconURL, url)
	}
	if c.specs[0].Icon == "" {
		t.Error("the emoji fallback must survive an icon upload")
	}
	if w := uploadTo(h, http.MethodPut, "/admin/specializations/"+slug+"/icon", coToken, "file", testPNG()); w.Code != http.StatusForbidden {
		t.Errorf("author upload: %d", w.Code)
	}
	if w := do(h, http.MethodDelete, "/admin/specializations/"+slug+"/icon", "", withCookie(adminToken)); w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	if c.specs[0].IconURL != "" || len(images.objects) != 0 {
		t.Fatalf("after delete: icon %q, objects %d", c.specs[0].IconURL, len(images.objects))
	}
}

func TestContentImageUpload(t *testing.T) {
	h, _, images := imagesFixture(t)
	w := uploadTo(h, http.MethodPost, "/admin/uploads/images", coToken, "file", testPNG())
	if w.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}
	url := decode[apigen.UploadedImage](t, w).URL
	if !strings.HasPrefix(url, "https://cdn.test/lessons/") {
		t.Fatalf("url = %q", url)
	}
	if _, ok := images.objects[url]; !ok {
		t.Fatal("image was not stored")
	}
	if w := uploadTo(h, http.MethodPost, "/admin/uploads/images", studentToken, "file", testPNG()); w.Code != http.StatusForbidden {
		t.Errorf("student upload: %d", w.Code)
	}
}

func TestDeletingCourseDropsItsImages(t *testing.T) {
	h, _, images := imagesFixture(t)
	icon := decode[apigen.UploadedImage](t, uploadTo(h, http.MethodPut, "/admin/courses/12/icon", coToken, "file", testPNG())).URL
	if w := uploadTo(h, http.MethodPut, "/admin/courses/12/cover", coToken, "file", testPNG()); w.Code != http.StatusNoContent {
		t.Fatalf("cover: %d", w.Code)
	}
	if w := do(h, http.MethodDelete, "/admin/courses/12", "", withCookie(ownerToken)); w.Code != http.StatusNoContent {
		t.Fatalf("delete course: %d %s", w.Code, w.Body)
	}
	if _, ok := images.objects[icon]; ok {
		t.Error("course icon outlived the course")
	}
	if len(images.objects) != 0 {
		t.Errorf("objects left in storage: %v", images.objects)
	}
}

func TestUploadStorageFailure(t *testing.T) {
	h, _, images := imagesFixture(t)
	images.putErr = errPutFailed
	if w := uploadTo(h, http.MethodPut, "/admin/courses/12/icon", coToken, "file", testPNG()); w.Code != http.StatusInternalServerError {
		t.Fatalf("put failure: %d %s", w.Code, w.Body)
	}
}
