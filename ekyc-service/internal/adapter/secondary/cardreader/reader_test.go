package cardreader

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/mrz"
)

// mrzTestLeftMargin matches the ~2% quiet-zone margin segmentCells trims
// from a rendered line's full width, so the synthetic text lines up with
// the fixed-pitch cell grid the production pipeline assumes.
const mrzTestLeftMargin = 4

func renderMRZLine(img *image.Gray, y int, text string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(mrzTestLeftMargin), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

// TestReadMRZ_SyntheticCard exercises the full pipeline end-to-end: PNG
// decode -> grayscale -> Otsu binarize -> MRZ band/line/cell segmentation
// -> glyph template matching -> TD1 field parsing -> checksum validation,
// against a synthetically rendered (but well-formed) card image.
func TestReadMRZ_SyntheticCard(t *testing.T) {
	// 30 chars * 7px advance (basicfont.Face7x13) + 4px margin each side.
	width, height := 218, 210
	img := image.NewGray(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = 255
	}

	docNumberField := "C1234567X"
	docCheck := testDigit(mrz.CheckDigit(docNumberField))
	optional1 := testPadTo("", 15)

	dob := "900115"
	dobCheck := testDigit(mrz.CheckDigit(dob))
	sex := "M"
	expiry := "300115"
	expiryCheck := testDigit(mrz.CheckDigit(expiry))
	nationality := "VNM"
	optional2 := testPadTo("", 11)

	composite := docNumberField + docCheck + optional1 + dob + dobCheck + expiry + expiryCheck + optional2
	compositeCheck := testDigit(mrz.CheckDigit(composite))

	line1 := "ID" + "VNM" + docNumberField + docCheck + optional1
	line2 := dob + dobCheck + sex + expiry + expiryCheck + nationality + optional2 + compositeCheck
	line3 := testPadTo("NGUYEN<<VAN<A", 30)

	// Mirrors locateMRZBand/splitLines: band = bottom 30% of a 210px-tall
	// image (rows 147-210), split into 3 equal 21px line strips.
	renderMRZLine(img, 163, line1)
	renderMRZLine(img, 184, line2)
	renderMRZLine(img, 205, line3)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode synthetic card: %v", err)
	}

	r := NewCardReader()
	result, err := r.ReadMRZ(context.Background(), buf.Bytes())
	if err != nil {
		t.Fatalf("ReadMRZ: %v", err)
	}

	t.Logf("line1=%q line2=%q line3=%q confidence=%.3f", result.Line1, result.Line2, result.Line3, result.Confidence)

	if result.Confidence < 0.9 {
		t.Errorf("confidence = %.3f, want >= 0.9 for a clean synthetic render", result.Confidence)
	}
	if !result.ChecksumValid {
		t.Errorf("expected checksum to validate on a correctly-generated MRZ (line1=%q line2=%q)", result.Line1, result.Line2)
	}
	if result.DocumentNumber != "C1234567X" {
		t.Errorf("DocumentNumber = %q, want C1234567X", result.DocumentNumber)
	}
	if result.Surname != "NGUYEN" || result.GivenNames != "VAN A" {
		t.Errorf("name = %q / %q, want NGUYEN / VAN A", result.Surname, result.GivenNames)
	}
}

func testDigit(n int) string {
	return string(rune('0' + n))
}

func testPadTo(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat("<", n-len(s))
}
