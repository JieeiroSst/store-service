package pyextract

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type pyFaceAnalyzer struct {
	client *Client
}

func NewFaceAnalyzer(client *Client) port.FaceAnalyzer {
	return &pyFaceAnalyzer{client: client}
}

const qualityScoreScale = 1000.0

func (a *pyFaceAnalyzer) Analyze(ctx context.Context, image []byte) (*model.FaceTemplate, error) {
	resp, err := a.client.call(ctx, map[string]any{
		"mode":      "face_embed",
		"image_b64": base64.StdEncoding.EncodeToString(image),
	})
	if err != nil {
		return nil, err
	}

	rawEmbedding, ok := resp["embedding"].([]any)
	if !ok {
		return nil, fmt.Errorf("pyextract: face_embed response missing 'embedding'")
	}
	embedding := make([]float64, len(rawEmbedding))
	for i, v := range rawEmbedding {
		f, _ := v.(float64)
		embedding[i] = f
	}

	areaRatio, _ := resp["quality_score"].(float64)

	return &model.FaceTemplate{
		Template:     encodeEmbedding(embedding),
		AlignedImage: image,
		QualityScore: areaRatio * qualityScoreScale,
	}, nil
}

func (a *pyFaceAnalyzer) AnalyzeFrames(ctx context.Context, frames [][]byte) (*model.FaceTemplate, error) {
	if len(frames) == 0 {
		return nil, port.ErrNoFaceDetected
	}

	var lastErr error
	for _, frame := range frames {
		tmpl, err := a.Analyze(ctx, frame)
		if err == nil {
			return tmpl, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (a *pyFaceAnalyzer) Compare(x, y []byte) float64 {
	ex := decodeEmbedding(x)
	ey := decodeEmbedding(y)
	if len(ex) != len(ey) || len(ex) == 0 {
		return 0
	}

	var sumSq float64
	for i := range ex {
		d := ex[i] - ey[i]
		sumSq += d * d
	}
	distance := math.Sqrt(sumSq)

	return 1.0 / (1.0 + distance)
}

func encodeEmbedding(embedding []float64) []byte {
	buf := make([]byte, 8*len(embedding))
	for i, f := range embedding {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(f))
	}
	return buf
}

func decodeEmbedding(buf []byte) []float64 {
	n := len(buf) / 8
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = math.Float64frombits(binary.LittleEndian.Uint64(buf[i*8:]))
	}
	return out
}
