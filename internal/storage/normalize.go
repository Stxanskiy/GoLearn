package storage

import (
	"bytes"
	"errors"
	"image"
	"image/png"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // decoder for WebP uploads
)

// ErrUndecodableImage is returned when a raster upload cannot be decoded for resizing.
var ErrUndecodableImage = errors.New("image cannot be decoded")

// Normalize downscales a raster upload to the stored side of its slot and re-encodes it as PNG.
// Vector art and images already within the limit are returned unchanged.
func Normalize(img Image, p Profile) (Image, error) {
	if p.StoredSide == 0 || img.ContentType == MimeSVG {
		return img, nil
	}
	w, h, ok := Dimensions(img.Data, img.ContentType)
	if !ok {
		return img, nil
	}
	if w <= p.StoredSide && h <= p.StoredSide && img.ContentType == MimePNG {
		return img, nil
	}

	src, _, err := image.Decode(bytes.NewReader(img.Data))
	if err != nil {
		return Image{}, ErrUndecodableImage
	}
	dw, dh := fit(w, h, p.StoredSide)
	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return Image{}, err
	}
	return Image{Data: buf.Bytes(), ContentType: MimePNG}, nil
}

// fit returns the target size that keeps the proportions within side.
func fit(w, h, side int) (int, int) {
	if w <= side && h <= side {
		return w, h
	}
	if w >= h {
		return side, max(1, h*side/w)
	}
	return max(1, w*side/h), side
}
