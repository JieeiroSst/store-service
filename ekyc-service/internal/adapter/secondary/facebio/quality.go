package facebio

func laplacianVariance(gray [][]uint8) float64 {
	h := len(gray)
	if h < 3 {
		return 0
	}
	w := len(gray[0])
	if w < 3 {
		return 0
	}

	responses := make([]float64, 0, (h-2)*(w-2))
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			lap := -4*int(gray[y][x]) +
				int(gray[y-1][x]) + int(gray[y+1][x]) +
				int(gray[y][x-1]) + int(gray[y][x+1])
			responses = append(responses, float64(lap))
		}
	}

	var mean float64
	for _, r := range responses {
		mean += r
	}
	mean /= float64(len(responses))

	var variance float64
	for _, r := range responses {
		d := r - mean
		variance += d * d
	}
	return variance / float64(len(responses))
}
