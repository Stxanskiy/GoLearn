package storage

import (
	"strings"
	"testing"
)

func TestDecodeImage(t *testing.T) {
	png := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...)
	gif := append([]byte("GIF89a"), make([]byte, 64)...)
	svg := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`)

	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"png", png, MimePNG},
		{"gif", gif, MimeGIF},
		{"svg", svg, MimeSVG},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			img, err := DecodeImage(c.data)
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
		if _, err := DecodeImage(data); err != ErrUnsupportedImage {
			t.Fatalf("err = %v, want ErrUnsupportedImage", err)
		}
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
