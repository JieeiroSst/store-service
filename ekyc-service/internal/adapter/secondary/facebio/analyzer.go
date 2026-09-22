package facebio

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type analyzer struct{}

func NewFaceAnalyzer() port.FaceAnalyzer {
	return &analyzer{}
}

func (a *analyzer) Analyze(ctx context.Context, image []byte) (*model.FaceTemplate, error) {
	img, err := decodeImage(image)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	pixels, rows, cols := grayscalePixels(img)
	det, err := detectLargestFace(pixels, rows, cols)
	if err != nil {
		return nil, err
	}

	face := cropAndResize(pixels, rows, cols, det.row, det.col, det.scale, faceSize)
	quality := laplacianVariance(face)
	template := computeLBPTemplate(face)

	return &model.FaceTemplate{
		Template:     template,
		AlignedImage: encodeGrayPNG(face),
		QualityScore: quality,
	}, nil
}

func (a *analyzer) AnalyzeFrames(ctx context.Context, frames [][]byte) (*model.FaceTemplate, error) {
	if len(frames) == 0 {
		return nil, port.ErrNoFaceDetected
	}
	if len(frames) == 1 {
		return a.Analyze(ctx, frames[0])
	}
	return a.analyzeFramesConcurrently(ctx, frames)
}

func (a *analyzer) Compare(x, y []byte) float64 {
	return compareLBP(x, y)
}
