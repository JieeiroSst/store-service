package pyextract

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/mrz"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type pyCardReader struct {
	client *Client
}

func NewCardReader(client *Client) port.CardReader {
	return &pyCardReader{client: client}
}

var mrzLineCharset = regexp.MustCompile(`^[A-Z0-9<]+$`)

type ocrLine struct {
	text       string
	confidence float64
}

func (r *pyCardReader) ReadMRZ(ctx context.Context, backImage []byte) (*model.MRZResult, error) {
	resp, err := r.client.call(ctx, map[string]any{
		"mode":      "ocr",
		"image_b64": base64.StdEncoding.EncodeToString(backImage),
	})
	if err != nil {
		return nil, err
	}

	rawLines, _ := resp["lines"].([]any)
	candidates := make([]ocrLine, 0, len(rawLines))
	for _, v := range rawLines {
		entry, ok := v.(map[string]any)
		if !ok {
			continue
		}
		text, _ := entry["text"].(string)
		conf, _ := entry["confidence"].(float64)
		candidates = append(candidates, ocrLine{text: normalizeMRZLine(text), confidence: conf})
	}

	mrzLines, avgConfidence := selectMRZLines(candidates)
	if len(mrzLines) != 3 {
		return nil, fmt.Errorf("pyextract: could not identify 3 MRZ-shaped lines among %d OCR results", len(candidates))
	}

	fields := mrz.ParseTD1(mrzLines[0], mrzLines[1], mrzLines[2])
	checksumValid := fields.DocNumberValid && fields.DOBValid && fields.ExpiryValid && fields.CompositeValid

	return &model.MRZResult{
		Line1:          mrzLines[0],
		Line2:          mrzLines[1],
		Line3:          mrzLines[2],
		DocumentNumber: fields.DocumentNumber,
		Surname:        fields.Surname,
		GivenNames:     fields.GivenNames,
		Nationality:    fields.Nationality,
		DateOfBirth:    fields.DateOfBirth,
		Sex:            fields.Sex,
		DateOfExpiry:   fields.DateOfExpiry,
		ChecksumValid:  checksumValid,
		Confidence:     avgConfidence,
	}, nil
}

func normalizeMRZLine(s string) string {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", "<"))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '<' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func selectMRZLines(candidates []ocrLine) ([]string, float64) {
	type scoredLine struct {
		text       string
		confidence float64
		index      int
		score      float64
	}
	var scored []scoredLine
	for i, c := range candidates {
		if len(c.text) < 20 || !mrzLineCharset.MatchString(c.text) {
			continue
		}
		lengthScore := 1.0 - absFloat(float64(len(c.text)-30))/30.0
		combined := 0.5*lengthScore + 0.5*c.confidence
		scored = append(scored, scoredLine{text: c.text, confidence: c.confidence, index: i, score: combined})
	}
	if len(scored) < 3 {
		return nil, 0
	}
	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
	top3 := scored[:3]
	sort.Slice(top3, func(i, j int) bool { return top3[i].index < top3[j].index })

	out := make([]string, 3)
	var confSum float64
	for i, s := range top3 {
		out[i] = padOrTruncate(s.text, 30)
		confSum += s.confidence
	}
	return out, confSum / 3
}

func padOrTruncate(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat("<", n-len(s))
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
