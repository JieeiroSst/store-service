package imaging

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/webp"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

type Composer struct{}

// NewComposer constructs the ImageComposer adapter.
func NewComposer() *Composer {
	return &Composer{}
}

var _ ports.ImageComposer = (*Composer)(nil)

func (c *Composer) Compose(ctx context.Context, sources []image.Image, layout domain.LayoutConfig) ([]byte, error) {
	if len(sources) == 0 {
		return nil, domain.ErrNoSources
	}

	cells, err := layout.ResolveCells(len(sources))
	if err != nil {
		return nil, err
	}

	bg, err := parseHexColor(layout.Background)
	if err != nil {
		return nil, fmt.Errorf("%w: background color: %v", domain.ErrInvalidLayout, err)
	}

	canvas := imaging.New(layout.Width, layout.Height, bg)

	for _, cell := range cells {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if cell.ImageIndex < 0 || cell.ImageIndex >= len(sources) {
			return nil, fmt.Errorf("%w: cell references image index %d, have %d sources", domain.ErrInvalidCellSpec, cell.ImageIndex, len(sources))
		}
		src := sources[cell.ImageIndex]

		x, y := cell.X, cell.Y
		var fitted *image.NRGBA
		switch layout.CellFit {
		case domain.FitContain:
			fitted = imaging.Fit(src, cell.Width, cell.Height, imaging.Lanczos)
			x += (cell.Width - fitted.Bounds().Dx()) / 2
			y += (cell.Height - fitted.Bounds().Dy()) / 2
		case domain.FitStretch:
			fitted = imaging.Resize(src, cell.Width, cell.Height, imaging.Lanczos)
		default: // FitCover
			fitted = imaging.Fill(src, cell.Width, cell.Height, imaging.Center, imaging.Lanczos)
		}

		canvas = imaging.Paste(canvas, fitted, image.Pt(x, y))
	}

	return encode(canvas, layout.Format, layout.Quality)
}

func encode(img image.Image, format domain.OutputFormat, quality int) ([]byte, error) {
	buf := new(bytes.Buffer)
	switch format {
	case domain.OutputFormatJPEG:
		if err := imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(quality)); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrComposeFailed, err)
		}
	case domain.OutputFormatPNG:
		if err := imaging.Encode(buf, img, imaging.PNG); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrComposeFailed, err)
		}
	case domain.OutputFormatWebP:
		if err := webp.Encode(buf, img, webp.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrComposeFailed, err)
		}
	default:
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedFormat, format)
	}
	return buf.Bytes(), nil
}

func parseHexColor(hex string) (color.NRGBA, error) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if hex == "" {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}, nil
	}
	if len(hex) != 6 {
		return color.NRGBA{}, fmt.Errorf("expected 6 hex digits, got %q", hex)
	}
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.NRGBA{}, err
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.NRGBA{}, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.NRGBA{}, err
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}
