package cardreader

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	glyphAlphabet  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<"
	templateWidth  = 7
	templateHeight = 13
)

var (
	glyphTemplates                            = buildGlyphTemplates()
	templateContentTop, templateContentBottom = inkRowBounds(glyphTemplates['0'])
)

func buildGlyphTemplates() map[rune][][]bool {
	templates := make(map[rune][][]bool, len(glyphAlphabet))
	for _, r := range glyphAlphabet {
		templates[r] = renderGlyph(r)
	}
	return templates
}

func inkRowBounds(bm [][]bool) (int, int) {
	top, bottom := len(bm), 0
	for y, row := range bm {
		for _, v := range row {
			if v {
				if y < top {
					top = y
				}
				if y+1 > bottom {
					bottom = y + 1
				}
				break
			}
		}
	}
	if top >= bottom {
		return 0, len(bm)
	}
	return top, bottom
}

func renderGlyph(r rune) [][]bool {
	img := image.NewGray(image.Rect(0, 0, templateWidth, templateHeight))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(11)},
	}
	d.DrawString(string(r))

	bitmap := make([][]bool, templateHeight)
	for y := 0; y < templateHeight; y++ {
		row := make([]bool, templateWidth)
		for x := 0; x < templateWidth; x++ {
			row[x] = img.GrayAt(x, y).Y < 128
		}
		bitmap[y] = row
	}
	return bitmap
}

func matchGlyph(cell [][]bool) (rune, float64) {
	bestRune := rune('<')
	bestScore := -1.0
	total := float64(templateWidth * templateHeight)

	for r, tmpl := range glyphTemplates {
		matches := 0
		for y := 0; y < templateHeight; y++ {
			for x := 0; x < templateWidth; x++ {
				if cell[y][x] == tmpl[y][x] {
					matches++
				}
			}
		}
		score := float64(matches) / total
		if score > bestScore {
			bestScore = score
			bestRune = r
		}
	}
	return bestRune, bestScore
}
