package http

import (
	"time"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/JIeeiroSst/cdn-service/internal/ports"
)

type presignRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	SizeBytes   int64  `json:"size_bytes" binding:"required,gt=0"`
}

type presignResponse struct {
	FileID    string    `json:"file_id"`
	ObjectKey string    `json:"object_key"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func toPresignResponse(pu *ports.PresignedUpload) presignResponse {
	return presignResponse{
		FileID:    pu.File.ID.String(),
		ObjectKey: pu.File.ObjectKey,
		UploadURL: pu.UploadURL,
		ExpiresAt: pu.ExpiresAt,
	}
}

type fileResponse struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toFileResponse(f model.File) fileResponse {
	return fileResponse{
		ID:          f.ID.String(),
		FileName:    f.FileName,
		ContentType: f.ContentType,
		SizeBytes:   f.SizeBytes,
		Status:      string(f.Status),
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

type listFilesResponse struct {
	Files []fileResponse `json:"files"`
}
