package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/JIeeiroSst/draw-image-service/internal/port"
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

const (
	decorationRatio = 0.28
	shadowOffset    = 8
	shadowBlur      = 8.0
	shadowOpacity   = 0.35
)

const supportedFormats = "jpeg, png, gif, bmp, tiff, webp"

var background = color.NRGBA{255, 255, 255, 255}

type processor struct {
	maxPixels int
}

func NewProcessor(cfg *config.Config) port.ImageProcessor {
	return &processor{maxPixels: cfg.Collage.MaxPixels}
}

func (p *processor) Decode(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidImage, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty file", domain.ErrInvalidImage)
	}

	header, _, err := image.DecodeConfig(bytes.NewReader(data))
	if errors.Is(err, image.ErrFormat) {
		return nil, fmt.Errorf("%w: unsupported format (supported: %s)", domain.ErrInvalidImage, supportedFormats)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidImage, err)
	}
	if header.Width <= 0 || header.Height <= 0 {
		return nil, fmt.Errorf("%w: invalid dimensions %dx%d", domain.ErrInvalidImage, header.Width, header.Height)
	}
	if int64(header.Width)*int64(header.Height) > int64(p.maxPixels) {
		return nil, fmt.Errorf("%w: %dx%d exceeds %d pixels",
			domain.ErrImageTooLarge, header.Width, header.Height, p.maxPixels)
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidImage, err)
	}
	return img, nil
}

func (p *processor) Thumbnail(img image.Image, width, height int) image.Image {
	return imaging.Thumbnail(img, width, height, imaging.CatmullRom)
}

func (p *processor) Shrink(img image.Image, maxWidth, maxHeight int) image.Image {
	b := img.Bounds()
	if b.Dx() <= maxWidth && b.Dy() <= maxHeight {
		return img
	}
	return contain(img, maxWidth, maxHeight)
}

func (p *processor) Compose(cells []image.Image, cellWidth, cellHeight, columns int) image.Image {
	rows := (len(cells) + columns - 1) / columns
	dst := imaging.New(cellWidth*columns, cellHeight*rows, background)
	for i, cell := range cells {
		pt := image.Pt(cellWidth*(i%columns), cellHeight*(i/columns))
		dst = imaging.Overlay(dst, cell, pt, 1)
	}
	return dst
}

func (p *processor) ComposeCenter(center image.Image, others []image.Image, width, height int) image.Image {
	decoSizes := make([]image.Point, len(others))
	for i, other := range others {
		decoSizes[i] = other.Bounds().Size()
	}
	mainRect, decoRects := layoutCenter(width, height, center.Bounds().Size(), decoSizes)

	dst := imaging.New(width, height, background)
	if mainRect.Empty() {
		return dst
	}
	dst = drawShadow(dst, mainRect)
	dst = imaging.Overlay(dst, resizeTo(center, mainRect.Size()), mainRect.Min, 1)
	for i, other := range others {
		if r := decoRects[i]; !r.Empty() {
			dst = imaging.Overlay(dst, resizeTo(other, r.Size()), r.Min, 1)
		}
	}
	return dst
}

func layoutCenter(width, height int, center image.Point, decos []image.Point) (image.Rectangle, []image.Rectangle) {
	rects := make([]image.Rectangle, len(decos))
	boxW := int(float64(width) * decorationRatio)
	boxH := int(float64(height) * decorationRatio)
	if center.X <= 0 || center.Y <= 0 || width <= boxW || height <= boxH {
		return image.Rectangle{}, rects
	}

	mw, mh := fitSize(center.X, center.Y, width-boxW, height-boxH)
	mainRect := image.Rect(0, 0, mw, mh).Add(image.Pt((width-mw)/2, (height-mh)/2))

	slots, alongX, alongY := perimeterSlots(len(decos))
	boxW = max(1, min(boxW, mw/(max(alongX, 1)+1)))
	boxH = max(1, min(boxH, mh/(max(alongY, 1)+1)))

	for i, d := range decos {
		if d.X <= 0 || d.Y <= 0 {
			continue
		}
		dw, dh := fitSize(d.X, d.Y, boxW, boxH)
		anchor := mainRect.Min.Add(image.Pt(
			int(slots[i][0]*float64(mw)),
			int(slots[i][1]*float64(mh)),
		))
		pt := image.Pt(
			clamp(anchor.X-dw/2, 0, width-dw),
			clamp(anchor.Y-dh/2, 0, height-dh),
		)
		rects[i] = image.Rect(0, 0, dw, dh).Add(pt)
	}
	return mainRect, rects
}

func perimeterSlots(n int) (slots [][2]float64, alongX, alongY int) {
	corners := [4][2]float64{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	onCorners := min(n, len(corners))
	slots = make([][2]float64, 0, n)
	slots = append(slots, corners[:onCorners]...)

	var perEdge [4]int
	for i := 0; i < n-onCorners; i++ {
		perEdge[i%4]++
	}
	for i := 0; i < n-onCorners; i++ {
		edge := i % 4
		t := float64(i/4+1) / float64(perEdge[edge]+1)
		switch edge {
		case 0:
			slots = append(slots, [2]float64{t, 0})
		case 1:
			slots = append(slots, [2]float64{t, 1})
		case 2:
			slots = append(slots, [2]float64{0, t})
		case 3:
			slots = append(slots, [2]float64{1, t})
		}
	}
	return slots, max(perEdge[0], perEdge[1]), max(perEdge[2], perEdge[3])
}

func resizeTo(img image.Image, size image.Point) image.Image {
	if img.Bounds().Size() == size {
		return img
	}
	return imaging.Resize(img, size.X, size.Y, imaging.Lanczos)
}

func (p *processor) EncodeJPEG(w io.Writer, img image.Image, quality int) error {
	return imaging.Encode(w, img, imaging.JPEG, imaging.JPEGQuality(quality))
}

func contain(img image.Image, w, h int) image.Image {
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 || w <= 0 || h <= 0 {
		return img
	}
	dw, dh := fitSize(b.Dx(), b.Dy(), w, h)
	if dw == b.Dx() && dh == b.Dy() {
		return img
	}
	return imaging.Resize(img, dw, dh, imaging.Lanczos)
}

func fitSize(srcW, srcH, w, h int) (int, int) {
	scale := math.Min(float64(w)/float64(srcW), float64(h)/float64(srcH))
	dw := min(w, max(1, int(math.Round(float64(srcW)*scale))))
	dh := min(h, max(1, int(math.Round(float64(srcH)*scale))))
	return dw, dh
}

func drawShadow(dst *image.NRGBA, r image.Rectangle) *image.NRGBA {
	pad := int(shadowBlur * 3)
	shadow := imaging.New(r.Dx()+2*pad, r.Dy()+2*pad, color.NRGBA{})
	shadow = imaging.Paste(shadow, imaging.New(r.Dx(), r.Dy(), color.NRGBA{0, 0, 0, 255}), image.Pt(pad, pad))
	shadow = imaging.Blur(shadow, shadowBlur)
	pt := r.Min.Add(image.Pt(shadowOffset-pad, shadowOffset-pad))
	return imaging.Overlay(dst, shadow, pt, shadowOpacity)
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
