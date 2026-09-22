package cardreader

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

func toGrayscale(img image.Image) [][]uint8 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	out := make([][]uint8, h)
	for y := 0; y < h; y++ {
		row := make([]uint8, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			// ITU-R BT.601 luma weights over the 16-bit RGBA components.
			lum := (299*r + 587*g + 114*b) / 1000
			row[x] = uint8(lum >> 8)
		}
		out[y] = row
	}
	return out
}

func otsuThreshold(gray [][]uint8) uint8 {
	var hist [256]int
	total := 0
	for _, row := range gray {
		for _, v := range row {
			hist[v]++
			total++
		}
	}
	if total == 0 {
		return 128
	}

	var sum float64
	for i, c := range hist {
		sum += float64(i) * float64(c)
	}

	var sumB, wB, maxVar float64
	threshold := 0
	for t := 0; t < 256; t++ {
		wB += float64(hist[t])
		if wB == 0 {
			continue
		}
		wF := float64(total) - wB
		if wF == 0 {
			break
		}
		sumB += float64(t) * float64(hist[t])
		mB := sumB / wB
		mF := (sum - sumB) / wF
		betweenVar := wB * wF * (mB - mF) * (mB - mF)
		if betweenVar > maxVar {
			maxVar = betweenVar
			threshold = t
		}
	}
	return uint8(threshold)
}

func binarize(gray [][]uint8, threshold uint8) [][]bool {
	out := make([][]bool, len(gray))
	for y, row := range gray {
		br := make([]bool, len(row))
		for x, v := range row {
			br[x] = v <= threshold
		}
		out[y] = br
	}
	return out
}
