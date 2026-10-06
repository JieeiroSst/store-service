package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/serpapi-service/internal/domain"
	"github.com/JIeeiroSst/serpapi-service/internal/port"
)

type ImageService struct {
	api     port.SerpAPI
	metrics port.Metrics
	limiter *Limiter
	now     func() time.Time
}

func NewImageService(api port.SerpAPI, metrics port.Metrics, limiter *Limiter) *ImageService {
	return &ImageService{api: api, metrics: metrics, limiter: limiter, now: time.Now}
}

func (s *ImageService) Upload(ctx context.Context, filename string, data []byte) (domain.UploadedImage, error) {
	img, err := domain.NewImageUpload(filename, data)
	if err != nil {
		return domain.UploadedImage{}, err
	}
	release, err := s.limiter.Acquire(ctx)
	if err != nil {
		return domain.UploadedImage{}, err
	}
	defer release()
	start := s.now()
	out, err := s.api.UploadImage(ctx, img)
	s.metrics.ObserveUpstream("image_upload", "", outcome(err), time.Since(start))
	if err != nil {
		return domain.UploadedImage{}, err
	}
	out.ExpiresAt = start.Add(domain.ImageIDTTL).UTC()
	return out, nil
}
