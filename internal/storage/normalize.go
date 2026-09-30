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
	// A slot that crops cannot take the shortcut unless the upload already has
	// the right proportions — otherwise a small square image was stored as is
	// and shown squashed into a 16:9 card.
	cropped := p.CropAspect && p.Aspect > 0 && !withinAspect(w, h, p.Aspect)
	if !cropped && w <= p.StoredSide && h <= p.StoredSide && img.ContentType == MimePNG {
		return img, nil
	}

	src, _, err := image.Decode(bytes.NewReader(img.Data))
	if err != nil {
		return Image{}, ErrUndecodableImage
	}
	if p.CropAspect && p.Aspect > 0 {
		src = cropToAspect(src, p.Aspect)
		b := src.Bounds()
		w, h = b.Dx(), b.Dy()
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

// withinAspect reports whether the image is already at the wanted proportions,
// give or take the rounding of a pixel or two.
func withinAspect(w, h int, aspect float64) bool {
	got := float64(w) / float64(h)
	return got >= aspect*(1-aspectTolerance) && got <= aspect*(1+aspectTolerance)
}

// cropToAspect takes the largest centred rectangle of the given width/height
// ratio. A cover is shown as a 16:9 banner whatever it was uploaded as, so the
// crop happens once here rather than being demanded of the author.
func cropToAspect(src image.Image, aspect float64) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	cw, ch := w, int(float64(w)/aspect+0.5)
	if ch > h {
		ch = h
		cw = int(float64(h)*aspect + 0.5)
	}
	if cw >= w && ch >= h {
		return src
	}
	rect := image.Rect(0, 0, cw, ch).Add(image.Pt(b.Min.X+(w-cw)/2, b.Min.Y+(h-ch)/2))
	type subImager interface {
		SubImage(image.Rectangle) image.Image
	}
	if si, ok := src.(subImager); ok {
		return si.SubImage(rect)
	}
	// Not every decoder returns a type that can subimage; copy the region out.
	dst := image.NewNRGBA(image.Rect(0, 0, cw, ch))
	draw.Draw(dst, dst.Bounds(), src, rect.Min, draw.Src)
	return dst
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
