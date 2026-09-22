package facebio

import (
	"context"
	"os"
	"testing"
)

// TestAnalyze_RealPhoto is a manual smoke test against a real photograph
// (not checked in - see the SAMPLE_FACE_IMAGE env var), exercising the
// full pipeline: JPEG decode -> pigo face detection -> crop/align ->
// LBP template -> quality score. It's skipped unless that env var points
// at a real image, since the repo doesn't vendor a photo of a person.
func TestAnalyze_RealPhoto(t *testing.T) {
	path := os.Getenv("SAMPLE_FACE_IMAGE")
	if path == "" {
		t.Skip("set SAMPLE_FACE_IMAGE to a path with a real face photo to run this smoke test")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	a := NewFaceAnalyzer()
	tmpl, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	t.Logf("template bytes=%d quality=%.2f", len(tmpl.Template), tmpl.QualityScore)
	if len(tmpl.Template) != lbpGridSize*lbpGridSize*lbpBins {
		t.Errorf("template length = %d, want %d", len(tmpl.Template), lbpGridSize*lbpGridSize*lbpBins)
	}

	self := a.Compare(tmpl.Template, tmpl.Template)
	if self < 0.999 {
		t.Errorf("self-comparison score = %.4f, want ~1.0", self)
	}
}
