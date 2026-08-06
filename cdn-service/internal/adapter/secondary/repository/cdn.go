package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/JIeeiroSst/cdn-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type cdnRepository struct {
	db *sql.DB
}

func NewCDNRepository(db *sql.DB) port.CDNRepository {
	return &cdnRepository{db: db}
}

func (r *cdnRepository) UploadFile(ctx context.Context, file model.File, meta []model.FileMetadata) (string, error) {
	lg := logger.WithContext(ctx)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		lg.Error("failed to begin transaction", zap.Error(err))
		return "", status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO files (id, filename, file_type, mime_type, size_bytes, storage_path, content_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, file.ID, file.Filename, file.FileType, file.MimeType, file.SizeBytes, file.StoragePath, file.ContentHash)
	if err != nil {
		lg.Error("failed to save file", zap.Error(err))
		return "", status.Errorf(codes.Internal, "failed to save file: %v", err)
	}

	if len(meta) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO file_metadata (id, file_id, metadata_key, metadata_value)
			VALUES ($1, $2, $3, $4)
		`)
		if err != nil {
			lg.Error("failed to prepare metadata statement", zap.Error(err))
			return "", status.Errorf(codes.Internal, "failed to prepare metadata statement: %v", err)
		}
		defer stmt.Close()

		for _, m := range meta {
			if _, err := stmt.ExecContext(ctx, m.ID, m.FileID, m.MetadataKey, m.MetadataValue); err != nil {
				lg.Error("failed to save metadata", zap.String("key", m.MetadataKey), zap.Error(err))
				return "", status.Errorf(codes.Internal, "failed to save metadata: %v", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		lg.Error("failed to commit transaction", zap.Error(err))
		return "", status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return file.ID, nil
}

func (r *cdnRepository) GetFile(ctx context.Context, fileID string) (*model.File, []model.FileMetadata, error) {
	lg := logger.WithContext(ctx)

	var file model.File
	err := r.db.QueryRowContext(ctx, `
		SELECT id, filename, file_type, mime_type, size_bytes, storage_path, content_hash
		FROM files WHERE id = $1
	`, fileID).Scan(&file.ID, &file.Filename, &file.FileType, &file.MimeType, &file.SizeBytes, &file.StoragePath, &file.ContentHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, status.Errorf(codes.NotFound, "file not found: %s", fileID)
		}
		lg.Error("failed to get file", zap.String("file_id", fileID), zap.Error(err))
		return nil, nil, status.Errorf(codes.Internal, "failed to get file: %v", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_id, metadata_key, metadata_value
		FROM file_metadata WHERE file_id = $1
	`, fileID)
	if err != nil {
		lg.Error("failed to get file metadata", zap.String("file_id", fileID), zap.Error(err))
		return nil, nil, status.Errorf(codes.Internal, "failed to get file metadata: %v", err)
	}
	defer rows.Close()

	var fileMeta []model.FileMetadata
	for rows.Next() {
		var meta model.FileMetadata
		if err := rows.Scan(&meta.ID, &meta.FileID, &meta.MetadataKey, &meta.MetadataValue); err != nil {
			return nil, nil, status.Errorf(codes.Internal, "failed to scan file metadata: %v", err)
		}
		fileMeta = append(fileMeta, meta)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, status.Errorf(codes.Internal, "failed to read file metadata: %v", err)
	}

	return &file, fileMeta, nil
}
