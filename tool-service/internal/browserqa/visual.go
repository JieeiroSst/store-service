package browserqa

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

func diffImages(base, cur []byte) (pct float64, sameSize bool, diffPNG []byte, err error) {
	a, err := png.Decode(bytes.NewReader(base))
	if err != nil {
		return 0, false, nil, err
	}
	b, err := png.Decode(bytes.NewReader(cur))
	if err != nil {
		return 0, false, nil, err
	}
	if a.Bounds().Size() != b.Bounds().Size() {
		return 100, false, nil, nil
	}
	const tolerance = 16
	bounds := b.Bounds()
	out := image.NewRGBA(bounds)
	total := bounds.Dx() * bounds.Dy()
	diff := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, _ := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()
			if absDiff(ar, br) > tolerance || absDiff(ag, bg) > tolerance || absDiff(ab, bb) > tolerance {
				diff++
				out.Set(x, y, color.RGBA{255, 0, 0, 255})
			} else {
				out.Set(x, y, color.RGBA{uint8(br >> 9), uint8(bg >> 9), uint8(bb >> 9), uint8(ba >> 8)})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return 0, true, nil, err
	}
	if total == 0 {
		return 0, true, buf.Bytes(), nil
	}
	return 100 * float64(diff) / float64(total), true, buf.Bytes(), nil
}

func absDiff(a, b uint32) uint32 {
	x, y := a>>8, b>>8
	if x > y {
		return x - y
	}
	return y - x
}
