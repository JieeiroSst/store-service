package imaging_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"testing"

	"github.com/JIeeiroSst/photo-service/internal/adapters/outbound/imaging"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

// solidImage generates a w x h image.RGBA filled with a single flat color,
// per the "ảnh tự sinh bằng image.NewRGBA" test requirement.
func solidImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	return img
}

func decodePNG(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode composer output: %v", err)
	}
	return img
}

func assertPixel(t *testing.T, img image.Image, x, y int, want color.RGBA) {
	t.Helper()
	r, g, b, a := img.At(x, y).RGBA()
	got := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	if got != want {
		t.Fatalf("pixel at (%d,%d): got %+v, want %+v", x, y, got, want)
	}
}

var (
	red    = color.RGBA{R: 255, A: 255}
	green  = color.RGBA{G: 255, A: 255}
	blue   = color.RGBA{B: 255, A: 255}
	yellow = color.RGBA{R: 255, G: 255, A: 255}
)

func TestCompose_Grid(t *testing.T) {
	// 2x2 grid, 220x220 canvas, padding=5, spacing=10 => each cell is exactly
	// 100x100, so 100x100 solid-color sources are placed with no resampling.
	layout := domain.LayoutConfig{
		Type:       domain.LayoutGrid,
		Cols:       2,
		Rows:       2,
		Width:      220,
		Height:     220,
		Padding:    5,
		Spacing:    10,
		Background: "#FFFFFF",
		CellFit:    domain.FitCover,
		Format:     domain.OutputFormatPNG,
	}
	sources := []image.Image{
		solidImage(100, 100, red),
		solidImage(100, 100, green),
		solidImage(100, 100, blue),
		solidImage(100, 100, yellow),
	}

	c := imaging.NewComposer()
	data, err := c.Compose(context.Background(), sources, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := decodePNG(t, data)
	if out.Bounds().Dx() != 220 || out.Bounds().Dy() != 220 {
		t.Fatalf("expected 220x220 canvas, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}

	assertPixel(t, out, 55, 55, red)      // cell (0,0)
	assertPixel(t, out, 165, 55, green)   // cell (1,0)
	assertPixel(t, out, 55, 165, blue)    // cell (0,1)
	assertPixel(t, out, 165, 165, yellow) // cell (1,1)
}

func TestCompose_Mosaic(t *testing.T) {
	// One hero cell (left half) + two stacked cells (right half), sized to
	// match their sources exactly so cover-fit performs no resampling.
	layout := domain.LayoutConfig{
		Type:       domain.LayoutMosaic,
		Width:      300,
		Height:     300,
		Background: "#FFFFFF",
		CellFit:    domain.FitCover,
		Format:     domain.OutputFormatPNG,
		Cells: []domain.CellSpec{
			{X: 0, Y: 0, Width: 150, Height: 300, ImageIndex: 0},
			{X: 150, Y: 0, Width: 150, Height: 150, ImageIndex: 1},
			{X: 150, Y: 150, Width: 150, Height: 150, ImageIndex: 2},
		},
	}
	sources := []image.Image{
		solidImage(150, 300, red),
		solidImage(150, 150, green),
		solidImage(150, 150, blue),
	}

	c := imaging.NewComposer()
	data, err := c.Compose(context.Background(), sources, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := decodePNG(t, data)
	if out.Bounds().Dx() != 300 || out.Bounds().Dy() != 300 {
		t.Fatalf("expected 300x300 canvas, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}

	assertPixel(t, out, 75, 150, red)   // hero cell
	assertPixel(t, out, 225, 75, green) // top-right cell
	assertPixel(t, out, 225, 225, blue) // bottom-right cell
}

func TestCompose_Freeform_OverlapPaintsLaterCellOnTop(t *testing.T) {
	// Two overlapping cells; the second (index 1) is expected to win in the
	// overlap region since freeform cells are painted in the order given.
	layout := domain.LayoutConfig{
		Type:       domain.LayoutFreeform,
		Width:      400,
		Height:     400,
		Background: "#FFFFFF",
		CellFit:    domain.FitCover,
		Format:     domain.OutputFormatPNG,
		Cells: []domain.CellSpec{
			{X: 0, Y: 0, Width: 250, Height: 250, ImageIndex: 0},
			{X: 150, Y: 150, Width: 250, Height: 250, ImageIndex: 1},
		},
	}
	sources := []image.Image{
		solidImage(250, 250, red),
		solidImage(250, 250, green),
	}

	c := imaging.NewComposer()
	data, err := c.Compose(context.Background(), sources, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := decodePNG(t, data)
	assertPixel(t, out, 50, 50, red)     // only inside cell 0
	assertPixel(t, out, 350, 350, green) // only inside cell 1
	assertPixel(t, out, 200, 200, green) // overlap: cell 1 painted last, wins
}

func TestCompose_GridSourceCountMismatch(t *testing.T) {
	layout := domain.LayoutConfig{
		Type: domain.LayoutGrid, Cols: 2, Rows: 2, Width: 200, Height: 200,
		Background: "#FFFFFF", CellFit: domain.FitCover, Format: domain.OutputFormatPNG,
	}
	sources := []image.Image{solidImage(10, 10, red), solidImage(10, 10, green), solidImage(10, 10, blue)} // 3, grid needs 4

	c := imaging.NewComposer()
	_, err := c.Compose(context.Background(), sources, layout)
	if !errors.Is(err, domain.ErrSourceCellMismatch) {
		t.Fatalf("expected ErrSourceCellMismatch for too few sources, got %v", err)
	}
}

func TestCompose_MosaicTooManySourcesForCells(t *testing.T) {
	layout := domain.LayoutConfig{
		Type:       domain.LayoutMosaic,
		Width:      200,
		Height:     200,
		Background: "#FFFFFF",
		CellFit:    domain.FitCover,
		Format:     domain.OutputFormatPNG,
		Cells: []domain.CellSpec{
			{X: 0, Y: 0, Width: 200, Height: 200, ImageIndex: 0},
		},
	}
	// two sources but only one cell defined
	sources := []image.Image{solidImage(10, 10, red), solidImage(10, 10, green)}

	c := imaging.NewComposer()
	_, err := c.Compose(context.Background(), sources, layout)
	if !errors.Is(err, domain.ErrSourceCellMismatch) {
		t.Fatalf("expected ErrSourceCellMismatch for too many sources, got %v", err)
	}
}

func TestCompose_NoSources(t *testing.T) {
	c := imaging.NewComposer()
	_, err := c.Compose(context.Background(), nil, domain.LayoutConfig{Type: domain.LayoutGrid, Width: 100, Height: 100})
	if !errors.Is(err, domain.ErrNoSources) {
		t.Fatalf("expected ErrNoSources, got %v", err)
	}
}
