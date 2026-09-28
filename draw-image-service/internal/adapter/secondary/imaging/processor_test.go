package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
)

var (
	red  = color.NRGBA{255, 0, 0, 255}
	blue = color.NRGBA{0, 0, 255, 255}
)

func solid(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func pngOf(t *testing.T, w, h int, c color.Color) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, solid(w, h, c)); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func newTestProcessor(maxPixels int) *processor {
	return NewProcessor(&config.Config{Collage: config.CollageConfig{MaxPixels: maxPixels}}).(*processor)
}

func isColor(t *testing.T, img image.Image, x, y int, want color.NRGBA) bool {
	t.Helper()
	got := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	near := func(a, b uint8) bool { d := int(a) - int(b); return d > -40 && d < 40 }
	return near(got.R, want.R) && near(got.G, want.G) && near(got.B, want.B)
}

func TestDecodeRejectsInvalidAndHugeImages(t *testing.T) {
	p := newTestProcessor(10_000)
	if _, err := p.Decode(strings.NewReader("not an image")); !errors.Is(err, domain.ErrInvalidImage) {
		t.Fatalf("err=%v, want ErrInvalidImage", err)
	}
	if _, err := p.Decode(pngOf(t, 200, 200, red)); !errors.Is(err, domain.ErrImageTooLarge) {
		t.Fatalf("err=%v, want ErrImageTooLarge", err)
	}
	img, err := p.Decode(pngOf(t, 50, 40, red))
	if err != nil || img.Bounds().Dx() != 50 || img.Bounds().Dy() != 40 {
		t.Fatalf("decode: %v %v", img, err)
	}
}

func TestThumbnailFillsCell(t *testing.T) {
	thumb := newTestProcessor(1_000_000).Thumbnail(solid(300, 200, red), 100, 100)
	if b := thumb.Bounds(); b.Dx() != 100 || b.Dy() != 100 {
		t.Fatalf("thumb size %v", b)
	}
}

func TestComposeGridPlacesCells(t *testing.T) {
	p := newTestProcessor(1_000_000)
	dst := p.Compose([]image.Image{solid(10, 10, red), solid(10, 10, red), solid(10, 10, blue)}, 10, 10, 2)
	if b := dst.Bounds(); b.Dx() != 20 || b.Dy() != 20 {
		t.Fatalf("collage size %v", b)
	}
	if !isColor(t, dst, 5, 15, blue) || !isColor(t, dst, 15, 5, red) {
		t.Fatal("cells misplaced")
	}

	var buf bytes.Buffer
	if err := p.EncodeJPEG(&buf, dst, 90); err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(&buf); err != nil {
		t.Fatalf("output is not a JPEG: %v", err)
	}
}

func TestComposeCenterFeaturesCenterAndOverlaysCorners(t *testing.T) {
	p := newTestProcessor(1_000_000)
	dst := p.ComposeCenter(solid(300, 400, red), []image.Image{solid(200, 100, blue)}, 400, 400)

	if b := dst.Bounds(); b.Dx() != 400 || b.Dy() != 400 {
		t.Fatalf("canvas %v", b)
	}
	if !isColor(t, dst, 200, 200, red) || !isColor(t, dst, 200, 70, red) || !isColor(t, dst, 200, 330, red) {
		t.Fatal("centre image not drawn large in the middle")
	}
	if !isColor(t, dst, 2, 200, color.NRGBA{255, 255, 255, 255}) {
		t.Fatal("background not white")
	}
	mainLeft := (400 - 216) / 2
	mainTop := (400 - 288) / 2
	if !isColor(t, dst, mainLeft+5, mainTop+5, blue) {
		t.Fatal("decoration does not overlap the centre's top-left corner")
	}
	if !isColor(t, dst, 400-mainLeft-10, 400-mainTop-10, red) {
		t.Fatal("unexpected decoration on bottom-right")
	}
}

func TestComposeCenterDecorationsDoNotOverlapEachOther(t *testing.T) {
	colors := []color.NRGBA{
		{0, 0, 255, 255}, {0, 255, 0, 255}, {255, 255, 0, 255}, {0, 255, 255, 255},
		{255, 0, 255, 255}, {0, 0, 0, 255}, {128, 128, 128, 255}, {255, 128, 0, 255},
	}
	others := make([]image.Image, len(colors))
	for i, c := range colors {
		others[i] = solid(100, 100, c)
	}
	dst := newTestProcessor(1_000_000).ComposeCenter(solid(300, 400, red), others, 400, 400)

	mainLeft, mainTop := (400-216)/2, (400-288)/2
	if !isColor(t, dst, mainLeft+50, mainTop, colors[0]) {
		t.Fatal("top-left decoration was covered by its neighbour")
	}
	if !isColor(t, dst, mainLeft+108-50, mainTop, colors[4]) {
		t.Fatal("top-centre decoration misplaced")
	}
	if !isColor(t, dst, 200, 200, red) {
		t.Fatal("centre image hidden by decorations")
	}
}
