package facebio

const (
	faceSize    = 64
	lbpGridSize = 8                      // 8x8 grid of cells over the aligned face
	lbpCellSize = faceSize / lbpGridSize // 8x8 pixels per cell
	lbpBins     = 256                    // one bin per possible 8-bit LBP code
)

func computeLBPTemplate(face [][]uint8) []byte {
	codes := make([][]uint8, faceSize)
	for y := range codes {
		codes[y] = make([]uint8, faceSize)
	}
	for y := 1; y < faceSize-1; y++ {
		for x := 1; x < faceSize-1; x++ {
			codes[y][x] = lbpCode(face, x, y)
		}
	}

	template := make([]byte, lbpGridSize*lbpGridSize*lbpBins)
	for gy := 0; gy < lbpGridSize; gy++ {
		for gx := 0; gx < lbpGridSize; gx++ {
			base := (gy*lbpGridSize + gx) * lbpBins
			for y := gy * lbpCellSize; y < (gy+1)*lbpCellSize; y++ {
				for x := gx * lbpCellSize; x < (gx+1)*lbpCellSize; x++ {
					template[base+int(codes[y][x])]++
				}
			}
		}
	}
	return template
}

func lbpCode(img [][]uint8, x, y int) uint8 {
	center := img[y][x]
	var code uint8
	// Clockwise from top-left, radius 1.
	neighbors := [8][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{1, 0},
		{1, 1}, {0, 1}, {-1, 1},
		{-1, 0},
	}
	for i, n := range neighbors {
		if img[y+n[1]][x+n[0]] >= center {
			code |= 1 << uint(i)
		}
	}
	return code
}

func compareLBP(a, b []byte) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var chiSq float64
	for i := range a {
		ai, bi := float64(a[i]), float64(b[i])
		denom := ai + bi
		if denom == 0 {
			continue
		}
		diff := ai - bi
		chiSq += (diff * diff) / denom
	}

	return 1.0 / (1.0 + chiSq/float64(len(a)))
}
