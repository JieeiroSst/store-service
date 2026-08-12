package ports

import (
	"context"
	"time"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/google/uuid"
)

type CreatePresignInput struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	UploadedBy  string
}

type PresignedUpload struct {
	File      model.File
	UploadURL string
	ExpiresAt time.Time
}

type ListFilesInput struct {
	Limit  int
	Offset int
}

type FileUsecase interface {
	CreatePresignedUpload(ctx context.Context, in CreatePresignInput) (*PresignedUpload, error)
	ConfirmUpload(ctx context.Context, fileID uuid.UUID) (*model.File, error)
	GetFile(ctx context.Context, fileID uuid.UUID) (*model.File, error)
	GetDownloadURL(ctx context.Context, fileID uuid.UUID, direct bool) (string, error)
	ListFiles(ctx context.Context, in ListFilesInput) ([]model.File, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID) error
}
