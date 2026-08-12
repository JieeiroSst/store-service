package domain

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/JIeeiroSst/cdn-service/internal/ports"
	"github.com/google/uuid"
)

var unsafeFileNameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type FileServiceOptions struct {
	MaxUploadSizeBytes int64
	PresignPutExpiry   time.Duration
	PresignGetExpiry   time.Duration
	EdgeCacheBaseURL   string
}

type FileService struct {
	repo    ports.FileRepository
	storage ports.ObjectStorage
	opts    FileServiceOptions
}

func NewFileService(repo ports.FileRepository, storage ports.ObjectStorage, opts FileServiceOptions) *FileService {
	return &FileService{repo: repo, storage: storage, opts: opts}
}

func (s *FileService) CreatePresignedUpload(ctx context.Context, in ports.CreatePresignInput) (*ports.PresignedUpload, error) {
	if in.FileName == "" || in.ContentType == "" {
		return nil, ErrInvalidFileInput
	}
	if in.SizeBytes <= 0 {
		return nil, ErrInvalidFileInput
	}
	if s.opts.MaxUploadSizeBytes > 0 && in.SizeBytes > s.opts.MaxUploadSizeBytes {
		return nil, ErrFileTooLarge
	}

	id := uuid.New()
	objectKey := buildObjectKey(id, in.FileName)

	uploadURL, err := s.storage.PresignPutURL(ctx, objectKey, s.opts.PresignPutExpiry)
	if err != nil {
		return nil, fmt.Errorf("presign put url: %w", err)
	}

	file := &model.File{
		ID:          id,
		ObjectKey:   objectKey,
		FileName:    in.FileName,
		ContentType: in.ContentType,
		SizeBytes:   in.SizeBytes,
		Status:      model.FileStatusPending,
		UploadedBy:  in.UploadedBy,
	}
	if err := s.repo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("create file metadata: %w", err)
	}

	return &ports.PresignedUpload{
		File:      *file,
		UploadURL: uploadURL,
		ExpiresAt: time.Now().Add(s.opts.PresignPutExpiry),
	}, nil
}

func (s *FileService) ConfirmUpload(ctx context.Context, fileID uuid.UUID) (*model.File, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}

	size, err := s.storage.StatObject(ctx, file.ObjectKey)
	if err != nil {
		file.Status = model.FileStatusFailed
		_ = s.repo.Update(ctx, file)
		return nil, fmt.Errorf("stat uploaded object: %w", err)
	}

	file.SizeBytes = size
	file.Status = model.FileStatusUploaded
	if err := s.repo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("update file metadata: %w", err)
	}

	return file, nil
}

func (s *FileService) GetFile(ctx context.Context, fileID uuid.UUID) (*model.File, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (s *FileService) GetDownloadURL(ctx context.Context, fileID uuid.UUID, direct bool) (string, error) {
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return "", err
	}
	if file.Status != model.FileStatusUploaded {
		return "", ErrInvalidStatus
	}

	if direct {
		return s.storage.PresignGetURL(ctx, file.ObjectKey, s.opts.PresignGetExpiry)
	}

	base := strings.TrimRight(s.opts.EdgeCacheBaseURL, "/")
	return fmt.Sprintf("%s/%s", base, file.ObjectKey), nil
}

func (s *FileService) ListFiles(ctx context.Context, in ports.ListFilesInput) ([]model.File, error) {
	limit := in.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *FileService) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return err
	}
	if err := s.storage.DeleteObject(ctx, file.ObjectKey); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return s.repo.Delete(ctx, fileID)
}

func buildObjectKey(id uuid.UUID, fileName string) string {
	safeName := unsafeFileNameChars.ReplaceAllString(filepath.Base(fileName), "_")
	return fmt.Sprintf("uploads/%s/%s", id.String(), safeName)
}
