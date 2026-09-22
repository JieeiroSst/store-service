package facebio

import "testing"

func syntheticFace(seed int) [][]uint8 {
	face := make([][]uint8, faceSize)
	for y := 0; y < faceSize; y++ {
		face[y] = make([]uint8, faceSize)
		for x := 0; x < faceSize; x++ {
			face[y][x] = uint8((x*7 + y*13 + seed*29) % 256)
		}
	}
	return face
}

func TestComputeLBPTemplate_Deterministic(t *testing.T) {
	face := syntheticFace(1)
	a := computeLBPTemplate(face)
	b := computeLBPTemplate(face)

	if len(a) != lbpGridSize*lbpGridSize*lbpBins {
		t.Fatalf("template length = %d, want %d", len(a), lbpGridSize*lbpGridSize*lbpBins)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("template not deterministic at byte %d: %d != %d", i, a[i], b[i])
		}
	}
}

func TestCompareLBP_IdenticalVsDifferent(t *testing.T) {
	a := computeLBPTemplate(syntheticFace(1))
	aAgain := computeLBPTemplate(syntheticFace(1))
	c := computeLBPTemplate(syntheticFace(99))

	identicalScore := compareLBP(a, aAgain)
	differentScore := compareLBP(a, c)

	if identicalScore < 0.999 {
		t.Errorf("identical templates scored %.4f, want ~1.0", identicalScore)
	}
	if differentScore >= identicalScore {
		t.Errorf("different templates scored %.4f, expected clearly lower than identical %.4f", differentScore, identicalScore)
	}
}

func TestLaplacianVariance_SharpVsBlurred(t *testing.T) {
	sharp := make([][]uint8, 32)
	blurred := make([][]uint8, 32)
	for y := 0; y < 32; y++ {
		sharp[y] = make([]uint8, 32)
		blurred[y] = make([]uint8, 32)
		for x := 0; x < 32; x++ {
			// Sharp: high-frequency checkerboard. Blurred: flat/near-constant.
			if (x+y)%2 == 0 {
				sharp[y][x] = 255
			}
			blurred[y][x] = 128
		}
	}

	sharpScore := laplacianVariance(sharp)
	blurredScore := laplacianVariance(blurred)

	if sharpScore <= blurredScore {
		t.Errorf("expected sharp image variance (%.2f) > blurred image variance (%.2f)", sharpScore, blurredScore)
	}
}
