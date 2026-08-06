package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/JIeeiroSst/cdn-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type cdnService struct {
	repo     port.CDNRepository
	baseHost config.BaseHostConfig
}

func NewCDNService(repo port.CDNRepository, cfg *config.Config) port.CDNUsecase {
	return &cdnService{repo: repo, baseHost: cfg.BaseHost}
}

func (s *cdnService) UploadFile(ctx context.Context, input model.UploadFileInput) (*model.FileResponse, error) {
	lg := logger.WithContext(ctx)

	if input.FileType != model.FileTypeImage && input.FileType != model.FileTypeVideo {
		lg.Error("invalid file type", zap.String("file_type", string(input.FileType)))
		return nil, status.Error(codes.InvalidArgument, "file type must be either IMAGE or VIDEO")
	}

	fileTypeDir := "images"
	if input.FileType == model.FileTypeVideo {
		fileTypeDir = "videos"
	}

	fileID := uuid.New().String()
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	storagePath := filepath.Join(fileTypeDir, datePath, fileID+filepath.Ext(input.Filename))

	fullPath := filepath.Join(s.baseHost.BaseDirUpload, storagePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		lg.Error("failed to create directory", zap.String("path", fullPath), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create directory: %v", err)
	}
	if err := os.WriteFile(fullPath, input.Content, 0644); err != nil {
		lg.Error("failed to write file", zap.String("path", fullPath), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to write file: %v", err)
	}

	file := model.File{
		ID:          fileID,
		Filename:    input.Filename,
		FileType:    input.FileType,
		MimeType:    input.MimeType,
		SizeBytes:   int64(len(input.Content)),
		StoragePath: storagePath,
		ContentHash: "hash-placeholder",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	meta := make([]model.FileMetadata, 0, len(input.Metadata))
	for key, value := range input.Metadata {
		meta = append(meta, model.FileMetadata{
			ID:            uuid.New().String(),
			FileID:        fileID,
			MetadataKey:   key,
			MetadataValue: value,
			CreatedAt:     now,
		})
	}

	if _, err := s.repo.UploadFile(ctx, file, meta); err != nil {
		lg.Error("failed to save file metadata", zap.String("file_id", fileID), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to save file metadata: %v", err)
	}

	return &model.FileResponse{
		FileID:    file.ID,
		Filename:  file.Filename,
		SizeBytes: file.SizeBytes,
		MimeType:  file.MimeType,
		FileType:  file.FileType,
		URL:       fmt.Sprintf("%s/v1/files/%s/content", s.baseHost.DominServiceURL, file.ID),
		CreatedAt: file.CreatedAt,
		Metadata:  input.Metadata,
	}, nil
}

func (s *cdnService) GetFile(ctx context.Context, fileID string) (*model.FileResponse, error) {
	lg := logger.WithContext(ctx)

	file, meta, err := s.repo.GetFile(ctx, fileID)
	if err != nil {
		lg.Error("failed to get file", zap.String("file_id", fileID), zap.Error(err))
		return nil, err
	}

	metadata := make(map[string]string, len(meta))
	for _, m := range meta {
		metadata[m.MetadataKey] = m.MetadataValue
	}

	return &model.FileResponse{
		FileID:      file.ID,
		Filename:    file.Filename,
		SizeBytes:   file.SizeBytes,
		MimeType:    file.MimeType,
		FileType:    file.FileType,
		URL:         fmt.Sprintf("%s/v1/files/%s/content", s.baseHost.DominServiceURL, file.ID),
		CreatedAt:   file.CreatedAt,
		Metadata:    metadata,
		StoragePath: file.StoragePath,
	}, nil
}
