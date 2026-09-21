package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

// rectPNG renders an opaque PNG of the given size.
func rectPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func TestDecodeImageContentProfile(t *testing.T) {
	gif := append([]byte("GIF89a"), make([]byte, 64)...)
	svg := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect/></svg>`)

	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"png", rectPNG(t, 8, 8), MimePNG},
		{"gif", gif, MimeGIF},
		{"svg", svg, MimeSVG},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			img, err := DecodeImage(c.data, ContentProfile)
			if err != nil {
				t.Fatalf("DecodeImage: %v", err)
			}
			if img.ContentType != c.want {
				t.Fatalf("content type = %q, want %q", img.ContentType, c.want)
			}
			if !strings.HasPrefix(img.Extension(), ".") {
				t.Fatalf("extension = %q", img.Extension())
			}
		})
	}
}

func TestDecodeImageRejectsOther(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("hello world, definitely not an image at all")} {
		if _, err := DecodeImage(data, ContentProfile); err != ErrUnsupportedImage {
			t.Fatalf("err = %v, want ErrUnsupportedImage", err)
		}
	}
}

func TestIconProfile(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want error
	}{
		{"square", rectPNG(t, 256, 256), nil},
		{"square svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect/></svg>`), nil},
		{"not square", rectPNG(t, 256, 200), ErrImageProportions},
		{"too small", rectPNG(t, 64, 64), ErrImageTooSmall},
		{"oversized is normalized, not rejected", rectPNG(t, 4096, 4096), nil},
		{"jpeg rejected", append([]byte("\xff\xd8\xff\xe0"), make([]byte, 64)...), ErrUnsupportedImage},
		{"gif rejected", append([]byte("GIF89a"), make([]byte, 64)...), ErrUnsupportedImage},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := DecodeImage(c.data, IconProfile); err != c.want {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
		})
	}
}

func TestCoverProfile(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want error
	}{
		{"16:9", rectPNG(t, 1280, 720), nil},
		{"16:9 svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1280 720"><rect/></svg>`), nil},
		{"square", rectPNG(t, 720, 720), ErrImageProportions},
		{"too small", rectPNG(t, 640, 360), ErrImageTooSmall},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := DecodeImage(c.data, CoverProfile); err != c.want {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
		})
	}
}

func TestDimensions(t *testing.T) {
	if w, h, ok := Dimensions(rectPNG(t, 40, 20), MimePNG); !ok || w != 40 || h != 20 {
		t.Fatalf("png = %d x %d, ok = %v", w, h, ok)
	}
	svg := []byte(`<svg width="64px" height="32px" xmlns="http://www.w3.org/2000/svg"></svg>`)
	if w, h, ok := Dimensions(svg, MimeSVG); !ok || w != 64 || h != 32 {
		t.Fatalf("svg = %d x %d, ok = %v", w, h, ok)
	}
	if _, _, ok := Dimensions([]byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), MimeSVG); ok {
		t.Fatal("svg without a size must be unreadable")
	}
}

func TestKeyOf(t *testing.T) {
	s := &Store{bucket: "golearn", publicURL: "http://localhost:9000/golearn"}
	if key, ok := s.keyOf("http://localhost:9000/golearn/icons/abc.png"); !ok || key != "icons/abc.png" {
		t.Fatalf("keyOf = %q, %v", key, ok)
	}
	if _, ok := s.keyOf("https://example.com/other.png"); ok {
		t.Fatal("foreign URL must not be owned")
	}
}

// WebP fixtures of a 300x200 canvas in the three RIFF flavours: VP8, VP8L and VP8X.
var webpFixtures = map[string]string{
	"lossy": "UklGRtgAAABXRUJQVlA4IMwAAADQEACdASosAcgAPpFIoU0lpCMiICgAsBIJaW7hdqlwAABPeCch77ZOQ99snIe+2TkP" +
		"fbJyHvtk5D32ygZeLk5D32ycj/nk5D32ych9C+l2vFych78AQD32ych77ZjLT2ych77ZOcRFych77ZOREDych77ZOQ+A" +
		"/S7Xi5OQ99uZfp7ZOQ99soGXi5OQ99scAAD+/14ov/+A3XxbTL//7nA/7nA/7nA/jbgE5RlBwKY1GhSz4uu/PW/PW/PW" +
		"/PW/PW/PW/PW/PW/PW/AAAA=",
	"lossless": "UklGRiQAAABXRUJQVlA4TBcAAAAvK8ExAAcQ/Y/+BwAU6f9/iuh/6v/XAAA=",
	"alpha": "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",
}

func TestWebPDimensions(t *testing.T) {
	for name, encoded := range webpFixtures {
		t.Run(name, func(t *testing.T) {
			data, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			if mime, err := sniff(data); err != nil || mime != MimeWebP {
				t.Fatalf("sniff = %q, %v", mime, err)
			}
			w, h, ok := Dimensions(data, MimeWebP)
			if !ok || w != 300 || h != 200 {
				t.Fatalf("size = %d x %d, ok = %v; want 300 x 200", w, h, ok)
			}
		})
	}
}

func TestNormalizeIcon(t *testing.T) {
	img, err := DecodeImage(rectPNG(t, 1024, 1024), IconProfile)
	if err != nil {
		t.Fatalf("DecodeImage: %v", err)
	}
	out, err := Normalize(img, IconProfile)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if out.ContentType != MimePNG {
		t.Fatalf("content type = %q, want %q", out.ContentType, MimePNG)
	}
	w, h, ok := Dimensions(out.Data, out.ContentType)
	if !ok || w != IconProfile.StoredSide || h != IconProfile.StoredSide {
		t.Fatalf("size = %d x %d, ok = %v; want %d square", w, h, ok, IconProfile.StoredSide)
	}
	if len(out.Data) >= len(img.Data) {
		t.Fatalf("normalized size = %d, want less than %d", len(out.Data), len(img.Data))
	}
}

func TestNormalizeWebPIcon(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(webpFixtures["lossless"])
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	out, err := Normalize(Image{Data: data, ContentType: MimeWebP}, IconProfile)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	// A 300x200 WebP is re-encoded as PNG and capped at the stored side.
	if out.ContentType != MimePNG {
		t.Fatalf("content type = %q, want %q", out.ContentType, MimePNG)
	}
	if w, h, ok := Dimensions(out.Data, out.ContentType); !ok || w != 256 || h != 170 {
		t.Fatalf("size = %d x %d, ok = %v; want 256 x 170", w, h, ok)
	}
}

func TestNormalizeKeepsSmallPNGAndSVG(t *testing.T) {
	small := Image{Data: rectPNG(t, 200, 200), ContentType: MimePNG}
	if out, err := Normalize(small, IconProfile); err != nil || !bytes.Equal(out.Data, small.Data) {
		t.Fatalf("small PNG must pass through: %v", err)
	}
	svg := Image{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"/>`), ContentType: MimeSVG}
	if out, err := Normalize(svg, IconProfile); err != nil || out.ContentType != MimeSVG {
		t.Fatalf("SVG must pass through: %v", err)
	}
}

// pngHeader builds a PNG carrying only an IHDR chunk — enough to read the declared size.
func pngHeader(w, h int) []byte {
	ihdr := make([]byte, 0, 17)
	ihdr = append(ihdr, "IHDR"...)
	ihdr = binary.BigEndian.AppendUint32(ihdr, uint32(w))
	ihdr = binary.BigEndian.AppendUint32(ihdr, uint32(h))
	ihdr = append(ihdr, 8, 6, 0, 0, 0)

	out := []byte("\x89PNG\r\n\x1a\n")
	out = binary.BigEndian.AppendUint32(out, uint32(len(ihdr)-4))
	out = append(out, ihdr...)
	return binary.BigEndian.AppendUint32(out, crc32.ChecksumIEEE(ihdr))
}

func TestDecodeImageRejectsOversizedIcon(t *testing.T) {
	side := IconProfile.MaxSide + 1
	if _, err := DecodeImage(pngHeader(side, side), IconProfile); err != ErrImageTooLarge {
		t.Fatalf("err = %v, want ErrImageTooLarge", err)
	}
	if _, err := DecodeImage(pngHeader(IconProfile.MaxSide, IconProfile.MaxSide), IconProfile); err != nil {
		t.Fatalf("icon at the limit rejected: %v", err)
	}
}
