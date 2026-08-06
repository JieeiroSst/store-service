package port

import (
	"context"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
)

type CDNRepository interface {
	UploadFile(ctx context.Context, file model.File, meta []model.FileMetadata) (string, error)
	GetFile(ctx context.Context, fileID string) (*model.File, []model.FileMetadata, error)
}
