package ports

import (
	"context"
	"image"
	"io"
	"time"

	"github.com/JIeeiroSst/photo-service/internal/domain"
)

type ImageComposer interface {
	Compose(ctx context.Context, sources []image.Image, layout domain.LayoutConfig) ([]byte, error)
}

type ImageDecoder interface {
	Decode(ctx context.Context, data []byte) (image.Image, error)
}

type ImageFetcher interface {
	Fetch(ctx context.Context, url string) (data []byte, contentType string, err error)
}

type ImageStorage interface {
	Put(ctx context.Context, key string, r io.Reader, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	PresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type JobRepository interface {
	Save(ctx context.Context, job *domain.CompositionJob) error
	FindByID(ctx context.Context, id string) (*domain.CompositionJob, error)
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}
