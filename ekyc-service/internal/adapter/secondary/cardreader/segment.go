package cardreader

const (
	mrzLines        = 3
	mrzCharsPerLine = 30
)

func locateMRZBand(gray [][]uint8) [][]uint8 {
	h := len(gray)
	if h == 0 {
		return gray
	}
	start := h - h*3/10
	if start < 0 {
		start = 0
	}
	return gray[start:]
}

func splitLines(band [][]uint8) [][][]uint8 {
	h := len(band)
	lineHeight := h / mrzLines
	lines := make([][][]uint8, mrzLines)
	for i := 0; i < mrzLines; i++ {
		start := i * lineHeight
		end := start + lineHeight
		if i == mrzLines-1 || end > h {
			end = h
		}
		lines[i] = band[start:end]
	}
	return lines
}

func cropToInkRows(bin [][]bool) [][]bool {
	top, bottom := len(bin), 0
	for y, row := range bin {
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
		return bin
	}
	return bin[top:bottom]
}

func segmentCells(bin [][]bool) [][][]bool {
	bin = cropToInkRows(bin)

	width := 0
	if len(bin) > 0 {
		width = len(bin[0])
	}
	margin := width / 50
	left, right := margin, width-margin
	if right <= left {
		left, right = 0, width
	}
	cellWidth := float64(right-left) / float64(mrzCharsPerLine)

	cells := make([][][]bool, mrzCharsPerLine)
	for i := 0; i < mrzCharsPerLine; i++ {
		x0 := left + int(float64(i)*cellWidth)
		x1 := left + int(float64(i+1)*cellWidth)
		if x1 <= x0 {
			x1 = x0 + 1
		}
		cells[i] = resampleCell(bin, x0, x1)
	}
	return cells
}

func resampleCell(bin [][]bool, x0, x1 int) [][]bool {
	h := len(bin)
	w := x1 - x0
	out := make([][]bool, templateHeight)
	for ty := 0; ty < templateHeight; ty++ {
		out[ty] = make([]bool, templateWidth)
	}
	if w <= 0 || h == 0 {
		return out
	}

	contentHeight := templateContentBottom - templateContentTop
	for ty := templateContentTop; ty < templateContentBottom; ty++ {
		srcY := scaleEndpointAnchored(ty-templateContentTop, contentHeight, h)
		srcRow := bin[srcY]
		for tx := 0; tx < templateWidth; tx++ {
			srcX := x0 + scaleEndpointAnchored(tx, templateWidth, w)
			if srcX < 0 || srcX >= len(srcRow) {
				continue
			}
			out[ty][tx] = srcRow[srcX]
		}
	}
	return out
}

func scaleEndpointAnchored(i, dstLen, srcLen int) int {
	if dstLen <= 1 || srcLen <= 1 {
		return 0
	}
	srcIdx := (i*(srcLen-1) + (dstLen-1)/2) / (dstLen - 1)
	if srcIdx >= srcLen {
		srcIdx = srcLen - 1
	}
	if srcIdx < 0 {
		srcIdx = 0
	}
	return srcIdx
}
