package cardreader

import (
	"context"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/mrz"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type reader struct{}

func NewCardReader() port.CardReader {
	return &reader{}
}

type lineResult struct {
	index      int
	text       string
	confidence float64
}

func (r *reader) ReadMRZ(ctx context.Context, backImage []byte) (*model.MRZResult, error) {
	img, err := decodeImage(backImage)
	if err != nil {
		return nil, fmt.Errorf("decode back image: %w", err)
	}
	gray := toGrayscale(img)
	band := locateMRZBand(gray)
	lines := splitLines(band)

	results := make(chan lineResult, mrzLines)
	for i, line := range lines {
		go func(i int, line [][]uint8) {
			results <- recognizeLine(i, line)
		}(i, line)
	}

	texts := make([]string, mrzLines)
	var totalConfidence float64
	for i := 0; i < mrzLines; i++ {
		res := <-results
		texts[res.index] = res.text
		totalConfidence += res.confidence
	}
	confidence := totalConfidence / float64(mrzLines)

	fields := mrz.ParseTD1(texts[0], texts[1], texts[2])
	checksumValid := fields.DocNumberValid && fields.DOBValid && fields.ExpiryValid && fields.CompositeValid

	return &model.MRZResult{
		Line1:          texts[0],
		Line2:          texts[1],
		Line3:          texts[2],
		DocumentNumber: fields.DocumentNumber,
		Surname:        fields.Surname,
		GivenNames:     fields.GivenNames,
		Nationality:    fields.Nationality,
		DateOfBirth:    fields.DateOfBirth,
		Sex:            fields.Sex,
		DateOfExpiry:   fields.DateOfExpiry,
		ChecksumValid:  checksumValid,
		Confidence:     confidence,
	}, nil
}

func recognizeLine(index int, line [][]uint8) lineResult {
	threshold := otsuThreshold(line)
	bin := binarize(line, threshold)
	cells := segmentCells(bin)

	var b strings.Builder
	var totalScore float64
	for _, cell := range cells {
		ch, score := matchGlyph(cell)
		b.WriteRune(ch)
		totalScore += score
	}
	return lineResult{index: index, text: b.String(), confidence: totalScore / float64(mrzCharsPerLine)}
}
