package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/cdn-service/dto"
	"github.com/JIeeiroSst/cdn-service/internal/repository"
	"github.com/JIeeiroSst/cdn-service/model"
	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CDNs interface {
	UploadFile(ctx context.Context, req dto.UploadFileRequest) (*dto.FileResponse, error)
	GetFile(ctx context.Context, fileID string) (*dto.FileResponse, error)
}

type CDNUsecase struct {
	Repos    *repository.Repositories
	BaseHost config.BaseHostConfig
	Cache    expire.CacheHelper
}

func NewCDNUsecase(deps *repository.Repositories,
	baseHost config.BaseHostConfig,
	cache expire.CacheHelper) *CDNUsecase {
	return &CDNUsecase{
		Repos:    deps,
		BaseHost: baseHost,
		Cache:    cache,
	}
}
func (u *CDNUsecase) UploadFile(ctx context.Context, req dto.UploadFileRequest) (*dto.FileResponse, error) {
	fileID := uuid.New().String()
	lg := logger.WithContext(ctx)

	if req.FileType != pb.FileType_FILE_TYPE_IMAGE && req.FileType != pb.FileType_FILE_TYPE_VIDEO {
		lg.Error("invalid file type", zap.String("file_type", req.FileType.String()))
		return nil, status.Error(codes.InvalidArgument, "file type must be either IMAGE or VIDEO")
	}

	fileTypeStr := "images"
	if req.FileType == pb.FileType_FILE_TYPE_VIDEO {
		fileTypeStr = "videos"
	}

	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	storagePath := filepath.Join(fileTypeStr, datePath, fileID+filepath.Ext(req.Filename))

	fullPath := filepath.Join(u.BaseHost.BaseDirUpload, storagePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		lg.Error("failed to create directory", zap.String("path", fullPath), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create directory: %v", err)
	}

	file := model.File{
		ID:          fileID,
		Filename:    req.Filename,
		FileType:    fileTypeStr,
		MimeType:    req.MimeType,
		SizeBytes:   int64(len(req.Content)),
		StoragePath: storagePath,
		ContentHash: "hash-placeholder",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsDeleted:   false,
	}

	meta := make([]model.FileMetadata, 0, len(req.Metadata))
	for key, value := range req.Metadata {
		meta = append(meta, model.FileMetadata{
			ID:            uuid.New().String(),
			FileID:        fileID,
			MetadataKey:   key,
			MetadataValue: value,
			CreatedAt:     time.Now(),
		})
	}

	fileID, err := u.Repos.CDN.UploadFile(ctx, file, meta)
	if err != nil {
		lg.Error("failed to save file metadata", zap.String("file_id", fileID), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to save file metadata: %v", err)
	}
	fileResp := &dto.FileResponse{
		FileID:    fileID,
		Filename:  file.Filename,
		SizeBytes: file.SizeBytes,
		MimeType:  file.MimeType,
		FileType:  file.FileType,
		Url:       fmt.Sprintf("%s/v1/files/%s/content", u.BaseHost.DominServiceURL, file.ID),
		CreatedAt: file.CreatedAt,
		Metadata:  req.Metadata,
	}

	return fileResp, nil
}

func (u *CDNUsecase) GetFile(ctx context.Context, fileID string) (*dto.FileResponse, error) {
	lg := logger.WithContext(ctx)
	lg.Info("get file", zap.String("file_id", fileID))
	file, meta, err := u.Repos.CDN.GetFile(ctx, fileID)
	if err == nil {
		lg.Info("get file success", zap.String("file_id", fileID))
		return nil, status.Errorf(codes.Internal, "failed to get file: %v", err)
	}

	metadata := make(map[string]string)
	for _, m := range meta {
		metadata[m.MetadataKey] = m.MetadataValue
	}

	fileResp := &dto.FileResponse{
		FileID:      file.ID,
		Filename:    file.Filename,
		SizeBytes:   file.SizeBytes,
		MimeType:    file.MimeType,
		FileType:    file.FileType,
		Url:         fmt.Sprintf("%s/v1/files/%s/content", u.BaseHost.DominServiceURL, file.ID),
		CreatedAt:   file.CreatedAt,
		Metadata:    metadata,
		StoragePath: file.StoragePath,
	}

	return fileResp, nil
}
