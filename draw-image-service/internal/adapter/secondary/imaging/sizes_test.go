package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

var white = color.NRGBA{255, 255, 255, 255}

var inputSizes = []struct{ w, h int }{
	{1, 1},
	{2, 3},
	{7, 5},
	{100, 100},
	{1, 2000},
	{2000, 1},
	{3, 1500},
	{1500, 3},
	{640, 480},
	{480, 640},
	{3000, 2000},
	{2000, 3000},
}

func encodePNG(t *testing.T, img image.Image) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func hasColor(img image.Image, want color.NRGBA) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			got := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			d := func(a, b uint8) int { return max(int(a)-int(b), int(b)-int(a)) }
			if d(got.R, want.R) < 40 && d(got.G, want.G) < 40 && d(got.B, want.B) < 40 {
				return true
			}
		}
	}
	return false
}

func TestEveryInputSizeWorksInEveryLayout(t *testing.T) {
	p := newTestProcessor(10_000_000)
	for _, s := range inputSizes {
		t.Run(fmt.Sprintf("%dx%d", s.w, s.h), func(t *testing.T) {
			img, err := p.Decode(pngOf(t, s.w, s.h, red))
			if err != nil {
				t.Fatal(err)
			}
			if b := img.Bounds(); b.Dx() != s.w || b.Dy() != s.h {
				t.Fatalf("decoded as %v", b)
			}

			thumb := p.Thumbnail(img, 100, 100)
			if b := thumb.Bounds(); b.Dx() != 100 || b.Dy() != 100 {
				t.Fatalf("thumbnail %v, want 100x100", b)
			}
			grid := p.Compose([]image.Image{thumb, thumb, thumb}, 100, 100, 2)
			if b := grid.Bounds(); b.Dx() != 200 || b.Dy() != 200 {
				t.Fatalf("grid %v", b)
			}
			if !isColor(t, grid, 50, 50, red) || !isColor(t, grid, 150, 150, white) {
				t.Fatal("grid cell or empty cell has the wrong colour")
			}

			shrunk := p.Shrink(img, 800, 800)
			sb := shrunk.Bounds()
			if sb.Dx() < 1 || sb.Dy() < 1 || sb.Dx() > 800 || sb.Dy() > 800 {
				t.Fatalf("shrunk to %v", sb)
			}
			if s.w <= 800 && s.h <= 800 && (sb.Dx() != s.w || sb.Dy() != s.h) {
				t.Fatalf("small image %dx%d must not be resized, got %v", s.w, s.h, sb)
			}

			asCenter := p.ComposeCenter(shrunk, []image.Image{solid(50, 50, blue)}, 400, 400)
			if b := asCenter.Bounds(); b.Dx() != 400 || b.Dy() != 400 {
				t.Fatalf("center canvas %v", b)
			}
			if !hasColor(asCenter, red) {
				t.Fatal("centre image missing from canvas")
			}

			asDecoration := p.ComposeCenter(solid(300, 300, blue), []image.Image{shrunk}, 400, 400)
			if b := asDecoration.Bounds(); b.Dx() != 400 || b.Dy() != 400 {
				t.Fatalf("decoration canvas %v", b)
			}
			if !hasColor(asDecoration, red) {
				t.Fatal("decoration missing from canvas")
			}
		})
	}
}

func TestFitSizeKeepsAspectAndNeverZero(t *testing.T) {
	tests := []struct{ srcW, srcH, w, h, wantW, wantH int }{
		{100, 100, 50, 50, 50, 50},
		{200, 100, 100, 100, 100, 50},
		{100, 200, 100, 100, 50, 100},
		{1, 1, 576, 576, 576, 576},
		{1, 5000, 288, 288, 1, 288},
		{5000, 1, 288, 288, 288, 1},
		{418, 557, 576, 576, 432, 576},
		{480, 360, 224, 224, 224, 168},
	}
	for _, tt := range tests {
		w, h := fitSize(tt.srcW, tt.srcH, tt.w, tt.h)
		if w != tt.wantW || h != tt.wantH {
			t.Errorf("fitSize(%dx%d into %dx%d) = %dx%d, want %dx%d",
				tt.srcW, tt.srcH, tt.w, tt.h, w, h, tt.wantW, tt.wantH)
		}
	}
}

func TestShrinkOnlyDownscales(t *testing.T) {
	p := newTestProcessor(10_000_000)
	small := solid(10, 20, red)
	if got := p.Shrink(small, 800, 800); got != image.Image(small) {
		t.Fatal("image that already fits must be returned unchanged")
	}
	if b := p.Shrink(solid(1600, 400, red), 800, 800).Bounds(); b.Dx() != 800 || b.Dy() != 200 {
		t.Fatalf("got %v, want 800x200", b)
	}
}

func TestTransparentImagesSitOnWhite(t *testing.T) {
	p := newTestProcessor(10_000_000)
	clear := image.NewNRGBA(image.Rect(0, 0, 50, 50))
	img, err := p.Decode(encodePNG(t, clear))
	if err != nil {
		t.Fatal(err)
	}

	grid := p.Compose([]image.Image{p.Thumbnail(img, 100, 100)}, 100, 100, 1)
	var buf bytes.Buffer
	if err := p.EncodeJPEG(&buf, grid, 95); err != nil {
		t.Fatal(err)
	}
	out, err := jpeg.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !isColor(t, out, 50, 50, white) {
		t.Fatal("transparent pixels must become white, not black, in the JPEG")
	}

	center := p.ComposeCenter(solid(300, 300, red), []image.Image{img}, 400, 400)
	mainTopLeft := (400 - 288) / 2
	if !isColor(t, center, mainTopLeft+5, mainTopLeft+5, red) {
		t.Fatal("transparent decoration must not hide the centre image")
	}
}

func TestDecodesCommonFormatsAndColorModels(t *testing.T) {
	p := newTestProcessor(10_000_000)
	rect := image.Rect(0, 0, 40, 30)

	gray := image.NewGray(rect)
	paletted := image.NewPaletted(rect, palette.Plan9)
	deep := image.NewNRGBA64(rect)

	inputs := map[string]func() (*bytes.Buffer, error){
		"png gray":     func() (*bytes.Buffer, error) { return encodePNG(t, gray), nil },
		"png paletted": func() (*bytes.Buffer, error) { return encodePNG(t, paletted), nil },
		"png 16-bit":   func() (*bytes.Buffer, error) { return encodePNG(t, deep), nil },
		"jpeg": func() (*bytes.Buffer, error) {
			var b bytes.Buffer
			return &b, jpeg.Encode(&b, solid(40, 30, red), nil)
		},
		"gif": func() (*bytes.Buffer, error) {
			var b bytes.Buffer
			return &b, gif.Encode(&b, solid(40, 30, red), nil)
		},
		"bmp": func() (*bytes.Buffer, error) {
			var b bytes.Buffer
			return &b, bmp.Encode(&b, solid(40, 30, red))
		},
		"tiff": func() (*bytes.Buffer, error) {
			var b bytes.Buffer
			return &b, tiff.Encode(&b, solid(40, 30, red), nil)
		},
	}
	for name, encode := range inputs {
		t.Run(name, func(t *testing.T) {
			data, err := encode()
			if err != nil {
				t.Fatal(err)
			}
			img, err := p.Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			if b := img.Bounds(); b.Dx() != 40 || b.Dy() != 30 {
				t.Fatalf("decoded as %v", b)
			}
			if b := p.Thumbnail(img, 100, 100).Bounds(); b.Dx() != 100 || b.Dy() != 100 {
				t.Fatalf("thumbnail %v", b)
			}
		})
	}
}

func TestDecodeAppliesExifOrientation(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, solid(40, 20, red), nil); err != nil {
		t.Fatal(err)
	}
	img, err := newTestProcessor(10_000_000).Decode(bytes.NewReader(withOrientation(buf.Bytes(), 6)))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 20 || b.Dy() != 40 {
		t.Fatalf("orientation 6 should rotate 40x20 to 20x40, got %v", b)
	}
}

func withOrientation(jpegData []byte, orientation byte) []byte {
	exif := []byte{
		'E', 'x', 'i', 'f', 0, 0,
		'M', 'M', 0, 42, 0, 0, 0, 8,
		0, 1,
		0x01, 0x12, 0, 3, 0, 0, 0, 1, 0, orientation, 0, 0,
		0, 0, 0, 0,
	}
	size := len(exif) + 2
	app1 := append([]byte{0xFF, 0xE1, byte(size >> 8), byte(size)}, exif...)
	out := append([]byte{}, jpegData[:2]...)
	out = append(out, app1...)
	return append(out, jpegData[2:]...)
}

func TestDecodeRejectsBadInputWithClearErrors(t *testing.T) {
	p := newTestProcessor(1_000_000)
	tests := []struct {
		name    string
		data    []byte
		want    error
		message string
	}{
		{"empty file", nil, domain.ErrInvalidImage, "empty file"},
		{"text file", []byte("hello world"), domain.ErrInvalidImage, "unsupported format"},
		{"heic-like", append([]byte{0, 0, 0, 24}, []byte("ftypheic")...), domain.ErrInvalidImage, "unsupported format"},
		{"truncated png", pngOf(t, 50, 50, red).Bytes()[:40], domain.ErrInvalidImage, ""},
		{"over pixel limit", pngOf(t, 1001, 1000, red).Bytes(), domain.ErrImageTooLarge, "1001x1000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Decode(bytes.NewReader(tt.data))
			if !errors.Is(err, tt.want) {
				t.Fatalf("err=%v, want %v", err, tt.want)
			}
			if !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("err=%q should mention %q", err, tt.message)
			}
		})
	}
}

func TestDecodeAcceptsExactlyMaxPixels(t *testing.T) {
	if _, err := newTestProcessor(1_000_000).Decode(pngOf(t, 1000, 1000, red)); err != nil {
		t.Fatalf("image at the limit must be accepted: %v", err)
	}
}
