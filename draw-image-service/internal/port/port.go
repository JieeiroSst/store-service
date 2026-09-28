package port

import (
	"context"
	"image"
	"io"

	"github.com/JIeeiroSst/draw-image-service/internal/domain"
)

type CollageService interface {
	Create(ctx context.Context, req domain.CollageRequest) (*domain.Collage, error)
	Get(ctx context.Context, id string) (*domain.Collage, error)
	Open(ctx context.Context, id string) (*domain.Collage, io.ReadCloser, error)
}

type ImageProcessor interface {
	Decode(r io.Reader) (image.Image, error)
	Thumbnail(img image.Image, width, height int) image.Image
	Shrink(img image.Image, maxWidth, maxHeight int) image.Image
	Compose(cells []image.Image, cellWidth, cellHeight, columns int) image.Image
	ComposeCenter(center image.Image, others []image.Image, width, height int) image.Image
	EncodeJPEG(w io.Writer, img image.Image, quality int) error
}

type ObjectStorage interface {
	Bucket() string
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type CollageRepository interface {
	Save(ctx context.Context, c *domain.Collage) error
	FindByID(ctx context.Context, id string) (*domain.Collage, error)
}
