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

// Upload contract violations, reported to the client as 415 or 413.
var (
	ErrUnsupportedImage = errors.New("unsupported image type")
	ErrImageTooLarge    = errors.New("image exceeds the slot limit")
	ErrImageTooSmall    = errors.New("image is smaller than the slot minimum")
	ErrImageProportions = errors.New("image proportions do not match the slot")
)

var extensions = map[string]string{
	MimePNG:  ".png",
	MimeJPEG: ".jpg",
	MimeWebP: ".webp",
	MimeGIF:  ".gif",
	MimeSVG:  ".svg",
}

// Profile is the upload contract of one image slot.
type Profile struct {
	Mimes      []string
	MaxBytes   int
	Aspect     float64 // required width/height; 0 disables the check
	MinSide    int     // shortest side in pixels; 0 disables the check
	MaxSide    int     // longest side in pixels; 0 disables the check
	StoredSide int     // longest side kept in storage; 0 stores the upload as is
	Hint       string  // contract text sent with the error
}

// Upload profiles of the image slots.
var (
	// IconProfile: square marker shown at 32-56 px, transparent background expected.
	IconProfile = Profile{
		Mimes:      []string{MimePNG, MimeWebP, MimeSVG},
		MaxBytes:   4 << 20,
		Aspect:     1,
		MinSide:    128,
		MaxSide:    4096,
		StoredSide: 256,
		Hint:       "icon must be PNG, WebP or SVG, square, 128-4096 px a side, max 4 MiB",
	}
	// CoverProfile: 16:9 banner of the course card and of the course page header.
	CoverProfile = Profile{
		Mimes:    []string{MimePNG, MimeJPEG, MimeWebP, MimeSVG},
		MaxBytes: 4 << 20,
		Aspect:   16.0 / 9.0,
		MinSide:  540,
		Hint:     "cover must be PNG, JPEG, WebP or SVG, 16:9, at least 960x540, max 4 MiB",
	}
	// ContentProfile: illustration inside lesson content, any proportions.
	ContentProfile = Profile{
		Mimes:    []string{MimePNG, MimeJPEG, MimeWebP, MimeGIF, MimeSVG},
		MaxBytes: 4 << 20,
		Hint:     "image must be PNG, JPEG, WebP, GIF or SVG, max 4 MiB",
	}
)

// aspectTolerance is the relative deviation allowed from the profile proportions.
const aspectTolerance = 0.02

// Image is validated upload bytes with the content type they were sniffed as.
type Image struct {
	Data        []byte
	ContentType string
}

func (i Image) Extension() string { return extensions[i.ContentType] }

func (i Image) Reader() io.Reader { return bytes.NewReader(i.Data) }

// DecodeImage sniffs the bytes and checks them against the slot contract.
func DecodeImage(data []byte, p Profile) (Image, error) {
	mime, err := sniff(data)
	if err != nil {
		return Image{}, err
	}
	if !accepts(p.Mimes, mime) {
		return Image{}, ErrUnsupportedImage
	}
	if p.MaxBytes > 0 && len(data) > p.MaxBytes {
		return Image{}, ErrImageTooLarge
	}
	if err := checkGeometry(data, mime, p); err != nil {
		return Image{}, err
	}
	return Image{Data: data, ContentType: mime}, nil
}

func sniff(data []byte) (string, error) {
	if len(data) == 0 {
		return "", ErrUnsupportedImage
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
		return "", ErrUnsupportedImage
	}
	return mime, nil
}

func accepts(mimes []string, mime string) bool {
	for _, allowed := range mimes {
		if allowed == mime {
			return true
		}
	}
	return false
}

// checkGeometry applies the size and proportion rules; unreadable dimensions pass.
func checkGeometry(data []byte, mime string, p Profile) error {
	if p.Aspect == 0 && p.MinSide == 0 && p.MaxSide == 0 {
		return nil
	}
	w, h, ok := Dimensions(data, mime)
	if !ok {
		return nil
	}
	// Vector art scales to any slot, so only its proportions are worth checking.
	if mime != MimeSVG {
		short, long := w, h
		if short > long {
			short, long = long, short
		}
		if p.MinSide > 0 && short < p.MinSide {
			return ErrImageTooSmall
		}
		if p.MaxSide > 0 && long > p.MaxSide {
			return ErrImageTooLarge
		}
	}
	if p.Aspect > 0 {
		got := float64(w) / float64(h)
		if got < p.Aspect*(1-aspectTolerance) || got > p.Aspect*(1+aspectTolerance) {
			return ErrImageProportions
		}
	}
	return nil
}

func looksLikeSVG(data []byte) bool {
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	return bytes.Contains(bytes.ToLower(head), []byte("<svg"))
}
