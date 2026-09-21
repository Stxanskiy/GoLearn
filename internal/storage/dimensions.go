package storage

import (
	"bytes"
	"encoding/binary"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"regexp"
	"strconv"
	"strings"
)

// Dimensions returns the pixel size of an upload; ok is false when it cannot be read.
func Dimensions(data []byte, mime string) (w, h int, ok bool) {
	switch mime {
	case MimeSVG:
		return svgSize(data)
	case MimeWebP:
		return webpSize(data)
	default:
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
			return 0, 0, false
		}
		return cfg.Width, cfg.Height, true
	}
}

// webpSize reads the canvas size out of the VP8, VP8L or VP8X chunk of a RIFF container.
func webpSize(data []byte) (w, h int, ok bool) {
	if len(data) < 30 || !bytes.HasPrefix(data, []byte("RIFF")) || !bytes.Equal(data[8:12], []byte("WEBP")) {
		return 0, 0, false
	}
	switch string(data[12:16]) {
	case "VP8X":
		return int(uint32(data[24])|uint32(data[25])<<8|uint32(data[26])<<16) + 1,
			int(uint32(data[27])|uint32(data[28])<<8|uint32(data[29])<<16) + 1, true
	case "VP8 ":
		if len(data) < 30 || !bytes.Equal(data[23:26], []byte{0x9d, 0x01, 0x2a}) {
			return 0, 0, false
		}
		return int(binary.LittleEndian.Uint16(data[26:28]) & 0x3fff),
			int(binary.LittleEndian.Uint16(data[28:30]) & 0x3fff), true
	case "VP8L":
		if data[20] != 0x2f {
			return 0, 0, false
		}
		bits := binary.LittleEndian.Uint32(data[21:25])
		return int(bits&0x3fff) + 1, int((bits>>14)&0x3fff) + 1, true
	}
	return 0, 0, false
}

var (
	svgRootRe = regexp.MustCompile(`(?is)<svg\b[^>]*>`)
	svgAttrRe = regexp.MustCompile(`(?is)\b(width|height|viewBox)\s*=\s*"([^"]*)"|\b(width|height|viewBox)\s*=\s*'([^']*)'`)
)

// svgSize reads the user-unit size of the root element from viewBox, or from width and height.
func svgSize(data []byte) (w, h int, ok bool) {
	root := svgRootRe.Find(data)
	if root == nil {
		return 0, 0, false
	}
	attrs := map[string]string{}
	for _, m := range svgAttrRe.FindAllSubmatch(root, -1) {
		name, value := string(m[1]), string(m[2])
		if name == "" {
			name, value = string(m[3]), string(m[4])
		}
		attrs[strings.ToLower(name)] = value
	}
	if box := strings.Fields(strings.ReplaceAll(attrs["viewbox"], ",", " ")); len(box) == 4 {
		bw, hasW := svgLength(box[2])
		bh, hasH := svgLength(box[3])
		if hasW && hasH {
			return bw, bh, true
		}
	}
	aw, hasW := svgLength(attrs["width"])
	ah, hasH := svgLength(attrs["height"])
	if hasW && hasH {
		return aw, ah, true
	}
	return 0, 0, false
}

// svgLength parses an SVG length, dropping the unit suffix; percentages are not sizes.
func svgLength(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasSuffix(raw, "%") {
		return 0, false
	}
	raw = strings.TrimRight(raw, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ ")
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return int(value), true
}
