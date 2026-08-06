package port

import (
	"context"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
)

type CDNUsecase interface {
	UploadFile(ctx context.Context, input model.UploadFileInput) (*model.FileResponse, error)
	GetFile(ctx context.Context, fileID string) (*model.FileResponse, error)
}
