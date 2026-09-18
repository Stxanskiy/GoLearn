package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/backendraz/golearn/internal/storage"
)

// fakeImages records what the handlers put into and remove from object storage.
type fakeImages struct {
	objects map[string]storage.Image // url → stored image
	deleted []string
	putErr  error
	next    int
}

func newFakeImages() *fakeImages {
	return &fakeImages{objects: map[string]storage.Image{}}
}

func (f *fakeImages) Put(_ context.Context, kind storage.Kind, img storage.Image) (string, error) {
	if f.putErr != nil {
		return "", f.putErr
	}
	f.next++
	url := fmt.Sprintf("https://cdn.test/%s/%d%s", kind, f.next, img.Extension())
	f.objects[url] = img
	return url, nil
}

func (f *fakeImages) Delete(_ context.Context, url string) error {
	if _, ok := f.objects[url]; !ok {
		return nil
	}
	delete(f.objects, url)
	f.deleted = append(f.deleted, url)
	return nil
}

var errPutFailed = errors.New("put failed")
