package port

import (
	"context"
	"io"

	"github.com/JIeeiroSst/video-service/internal/domain/model"
)

type VideoStorage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (int64, error)
	ReadAt(ctx context.Context, key string, size int64, p []byte, off int64) (int, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
}

type VideoRepository interface {
	Save(ctx context.Context, v model.Video) error
	Get(ctx context.Context, id string) (*model.Video, error)
	List(ctx context.Context) ([]model.Video, error)
	Delete(ctx context.Context, id string) error
}

type ViewCounter interface {
	Incr(ctx context.Context, id, viewer string) error
	Counts(ctx context.Context, ids []string) (map[string]int64, error)
	Delete(ctx context.Context, id string) error
}

type JobQueue interface {
	Enqueue(ctx context.Context, id string) error
	Consume(ctx context.Context, handle func(ctx context.Context, id string) error) error
}

type InvalidationBus interface {
	Publish(ctx context.Context, id string) error
	Subscribe(ctx context.Context, fn func(id string)) error
}

type TranscodeResult struct {
	Duration float64
	Width    int
	Height   int
}

type Transcoder interface {
	Transcode(ctx context.Context, src, outDir string) (*TranscodeResult, error)
}
