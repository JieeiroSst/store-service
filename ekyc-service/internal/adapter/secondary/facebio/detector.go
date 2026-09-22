package facebio

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"

	pigo "github.com/esimov/pigo/core"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

//go:embed cascade/facefinder
var facefinderCascade []byte

var classifier = mustUnpackClassifier()

func mustUnpackClassifier() *pigo.Pigo {
	p := pigo.NewPigo()
	c, err := p.Unpack(facefinderCascade)
	if err != nil {
		panic(fmt.Sprintf("facebio: unpack embedded cascade: %v", err))
	}
	return c
}

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

func grayscalePixels(img image.Image) (pixels []uint8, rows, cols int) {
	bounds := img.Bounds()
	cols, rows = bounds.Dx(), bounds.Dy()
	pixels = pigo.RgbToGrayscale(img)
	return
}

type detection struct {
	row, col, scale int
	score           float32
}

func detectLargestFace(pixels []uint8, rows, cols int) (*detection, error) {
	minSize := rows / 10
	if cols/10 < minSize {
		minSize = cols / 10
	}
	if minSize < 20 {
		minSize = 20
	}
	maxSize := rows
	if cols > maxSize {
		maxSize = cols
	}

	cParams := pigo.CascadeParams{
		MinSize:     minSize,
		MaxSize:     maxSize,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	dets := classifier.RunCascade(cParams, 0.0)
	dets = classifier.ClusterDetections(dets, 0.2)
	if len(dets) == 0 {
		return nil, port.ErrNoFaceDetected
	}

	best := dets[0]
	for _, d := range dets[1:] {
		if d.Q > best.Q {
			best = d
		}
	}
	return &detection{row: best.Row, col: best.Col, scale: best.Scale, score: best.Q}, nil
}

func cropAndResize(pixels []uint8, rows, cols, row, col, scale, size int) [][]uint8 {
	half := scale / 2
	x0, y0 := col-half, row-half

	out := make([][]uint8, size)
	for ty := 0; ty < size; ty++ {
		out[ty] = make([]uint8, size)
		srcY := y0 + ty*scale/size
		if srcY < 0 {
			srcY = 0
		} else if srcY >= rows {
			srcY = rows - 1
		}
		for tx := 0; tx < size; tx++ {
			srcX := x0 + tx*scale/size
			if srcX < 0 {
				srcX = 0
			} else if srcX >= cols {
				srcX = cols - 1
			}
			out[ty][tx] = pixels[srcY*cols+srcX]
		}
	}
	return out
}

// encodeGrayPNG serializes an aligned face patch for storage as an audit
// trail image alongside its biometric template.
func encodeGrayPNG(face [][]uint8) []byte {
	size := len(face)
	img := image.NewGray(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetGray(x, y, color.Gray{Y: face[y][x]})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
