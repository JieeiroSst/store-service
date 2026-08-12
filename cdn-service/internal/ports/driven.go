package ports

import (
	"context"
	"time"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/google/uuid"
)

type FileRepository interface {
	Create(ctx context.Context, file *model.File) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.File, error)
	Update(ctx context.Context, file *model.File) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]model.File, error)
}

type ObjectStorage interface {
	PresignPutURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	PresignGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	StatObject(ctx context.Context, objectKey string) (sizeBytes int64, err error)
	DeleteObject(ctx context.Context, objectKey string) error
}
