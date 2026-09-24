package port

import (
	"context"
	"io"
	"time"

	"github.com/JIeeiroSst/upload-service/internal/domain/model"
)

type Upload struct {
	ReceiverID string
	FileName   string
	Body       io.Reader
}

type Download struct {
	File *model.File
	Body io.ReadCloser
}

type FileUsecase interface {
	Create(ctx context.Context, in Upload) (*model.File, error)
	Replace(ctx context.Context, id string, in Upload) (*model.File, error)
	Get(ctx context.Context, id string) (*model.File, error)
	Open(ctx context.Context, id string) (*Download, error)
	List(ctx context.Context, receiverID string, limit, offset int) ([]model.File, int64, error)
	Delete(ctx context.Context, id string) error
}

type MetadataRepository interface {
	Insert(ctx context.Context, f *model.File) error
	Get(ctx context.Context, id string) (*model.File, error)
	List(ctx context.Context, receiverID string, limit, offset int) ([]model.File, int64, error)
	ReplaceContent(ctx context.Context, f *model.File) error
	Delete(ctx context.Context, id string) error
}

type ObjectStore interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type TokenValidator interface {
	Validate(ctx context.Context, token string) (userID string, err error)
}

type Clock interface{ Now() time.Time }
