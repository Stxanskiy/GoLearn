package storage

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
)

const (
	MimePNG  = "image/png"
	MimeJPEG = "image/jpeg"
	MimeWebP = "image/webp"
	MimeGIF  = "image/gif"
	MimeSVG  = "image/svg+xml"
)

// ErrUnsupportedImage is returned for bytes that are not an allowed image type.
var ErrUnsupportedImage = errors.New("unsupported image type")

var extensions = map[string]string{
	MimePNG:  ".png",
	MimeJPEG: ".jpg",
	MimeWebP: ".webp",
	MimeGIF:  ".gif",
	MimeSVG:  ".svg",
}

// Image is validated upload bytes with the content type they were sniffed as.
type Image struct {
	Data        []byte
	ContentType string
}

func (i Image) Extension() string { return extensions[i.ContentType] }

func (i Image) Reader() io.Reader { return bytes.NewReader(i.Data) }

// DecodeImage sniffs the bytes and accepts only the image types the catalogue renders.
func DecodeImage(data []byte) (Image, error) {
	if len(data) == 0 {
		return Image{}, ErrUnsupportedImage
	}
	mime := http.DetectContentType(data)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if mime == "text/xml" || mime == "text/plain" {
		if looksLikeSVG(data) {
			mime = MimeSVG
		}
	}
	if _, ok := extensions[mime]; !ok {
		return Image{}, ErrUnsupportedImage
	}
	return Image{Data: data, ContentType: mime}, nil
}

func looksLikeSVG(data []byte) bool {
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	return bytes.Contains(bytes.ToLower(head), []byte("<svg"))
}
