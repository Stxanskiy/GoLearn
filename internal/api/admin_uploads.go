package api

import (
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/storage"
)

const maxImageBytes = 4 << 20

func (a *API) adminUploadCourseIcon(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok || !a.checkEditable(w, r, course) {
		return
	}
	url, ok := a.storeUpload(w, r, storage.KindIcon, storage.IconProfile)
	if !ok {
		return
	}
	a.dropStored(r, course.module.IconURL)
	if err := a.Modules.SetIcon(r.Context(), course.module.ID, url); err != nil {
		a.internalError(w, "admin: upload course icon", err)
		return
	}
	writeJSON(w, http.StatusOK, apigen.UploadedImage{URL: url})
}

func (a *API) adminDeleteCourseIcon(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok || !a.checkEditable(w, r, course) {
		return
	}
	a.dropStored(r, course.module.IconURL)
	if err := a.Modules.SetIcon(r.Context(), course.module.ID, ""); err != nil {
		a.internalError(w, "admin: delete course icon", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminUploadSpecIcon(w http.ResponseWriter, r *http.Request) {
	spec, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	url, ok := a.storeUpload(w, r, storage.KindIcon, storage.IconProfile)
	if !ok {
		return
	}
	a.dropStored(r, spec.IconURL)
	if err := a.Specs.SetIcon(r.Context(), spec.Slug, url); err != nil {
		a.internalError(w, "admin: upload specialization icon", err)
		return
	}
	writeJSON(w, http.StatusOK, apigen.UploadedImage{URL: url})
}

func (a *API) adminDeleteSpecIcon(w http.ResponseWriter, r *http.Request) {
	spec, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	a.dropStored(r, spec.IconURL)
	if err := a.Specs.SetIcon(r.Context(), spec.Slug, ""); err != nil {
		a.internalError(w, "admin: delete specialization icon", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminUploadImage(w http.ResponseWriter, r *http.Request) {
	url, ok := a.storeUpload(w, r, storage.KindLesson, storage.ContentProfile)
	if !ok {
		return
	}
	writeJSON(w, http.StatusCreated, apigen.UploadedImage{URL: url})
}

// storeUpload reads the multipart image and puts it into object storage.
func (a *API) storeUpload(w http.ResponseWriter, r *http.Request, kind storage.Kind, profile storage.Profile) (string, bool) {
	if a.Images == nil {
		writeError(w, http.StatusServiceUnavailable, codeStorageDisabled, "object storage is not configured")
		return "", false
	}
	img, ok := readImageUpload(w, r, profile)
	if !ok {
		return "", false
	}
	url, err := a.Images.Put(r.Context(), kind, img)
	if err != nil {
		a.internalError(w, "admin: store image", err)
		return "", false
	}
	return url, true
}

// dropStored removes a replaced object; a failure only costs disk space.
func (a *API) dropStored(r *http.Request, url string) {
	if a.Images == nil || url == "" {
		return
	}
	if err := a.Images.Delete(r.Context(), url); err != nil {
		a.log.Warn("admin: delete stored image", "url", url, "error", err)
	}
}

// readImageUpload reads the multipart "file" part, writing the error response itself when it returns false.
func readImageUpload(w http.ResponseWriter, r *http.Request, profile storage.Profile) (storage.Image, bool) {
	data, ok := readUploadBytes(w, r)
	if !ok {
		return storage.Image{}, false
	}
	img, err := storage.DecodeImage(data, profile)
	if err != nil {
		writeImageContractError(w, err, profile)
		return storage.Image{}, false
	}
	img, err = storage.Normalize(img, profile)
	if err != nil {
		writeImageContractError(w, err, profile)
		return storage.Image{}, false
	}
	return img, true
}

// writeImageContractError reports a rejected upload with the slot contract as the message.
func writeImageContractError(w http.ResponseWriter, err error, profile storage.Profile) {
	if errors.Is(err, storage.ErrImageTooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, profile.Hint)
		return
	}
	writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMedia, profile.Hint)
}

// readUploadBytes returns the raw bytes of the multipart "file" part.
func readUploadBytes(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	if ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || ct != "multipart/form-data" {
		writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMedia, "expected multipart/form-data body")
		return nil, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxImageBytes+64<<10)
	mr, err := r.MultipartReader()
	if err != nil {
		writeValidation(w, map[string]string{"file": fieldRequired})
		return nil, false
	}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, uploadReadError(w, err)
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, maxImageBytes+1))
		if err != nil {
			return nil, uploadReadError(w, err)
		}
		if len(data) > maxImageBytes {
			writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, "image exceeds 4 MiB")
			return nil, false
		}
		return data, true
	}
	writeValidation(w, map[string]string{"file": fieldRequired})
	return nil, false
}

func uploadReadError(w http.ResponseWriter, err error) bool {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, "image exceeds 4 MiB")
		return false
	}
	writeValidation(w, map[string]string{"file": fieldInvalidFormat})
	return false
}

// storeCover puts the cover into object storage, falling back to an inline data URI.
func (a *API) storeCover(w http.ResponseWriter, r *http.Request) (string, bool) {
	img, ok := readImageUpload(w, r, storage.CoverProfile)
	if !ok {
		return "", false
	}
	if a.Images == nil {
		return "data:" + img.ContentType + ";base64," + base64.StdEncoding.EncodeToString(img.Data), true
	}
	url, err := a.Images.Put(r.Context(), storage.KindCover, img)
	if err != nil {
		a.internalError(w, "admin: store cover", err)
		return "", false
	}
	return url, true
}
