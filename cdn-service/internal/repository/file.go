package repository

import (
	"context"
	"database/sql"

	"github.com/JIeeiroSst/cdn-service/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CDN interface {
	UploadFile(ctx context.Context, file model.File, meta []model.FileMetadata) (string, error)
	GetFile(ctx context.Context, fileID string) (*model.File, []model.FileMetadata, error)
}

type CdnRepository struct {
	db *sql.DB
}

func NewCdnRepository(db *sql.DB) *CdnRepository {
	return &CdnRepository{
		db: db,
	}
}

func (r *CdnRepository) UploadFile(ctx context.Context, file model.File, meta []model.FileMetadata) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "nil", status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO files (id, filename, file_type, mime_type, size_bytes, storage_path, content_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, file.ID, file.Filename, file.FileType, file.MimeType, file.ContentHash, file.StoragePath, file.ContentHash)

	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to save file metadata: %v", err)
	}

	if len(meta) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO file_metadata (id, file_id, metadata_key, metadata_value)
			VALUES ($1, $2, $3, $4)
		`)
		if err != nil {
			return "", status.Errorf(codes.Internal, "failed to prepare metadata statement: %v", err)
		}
		defer stmt.Close()

		for key, value := range meta {
			metadataID := uuid.New().String()
			if _, err := stmt.ExecContext(ctx, metadataID, file.ID, key, value); err != nil {
				return "", status.Errorf(codes.Internal, "failed to save metadata: %v", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return file.ID, nil
}

func (r *CdnRepository) GetFile(ctx context.Context, fileID string) (*model.File, []model.FileMetadata, error) {
	var (
		file     model.File
		fileMeta []model.FileMetadata
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT id, filename, file_type, mime_type, size_bytes, storage_path, content_hash
		FROM files WHERE id = $1
	`, fileID).Scan(&file.ID, &file.Filename, &file.FileType, &file.MimeType, &file.SizeBytes, &file.StoragePath, &file.ContentHash)

	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "failed to get file: %v", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_id, metadata_key, metadata_value
		FROM file_metadata WHERE file_id = $1
	`, fileID)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "failed to get file metadata: %v", err)
	}

	for rows.Next() {
		var meta model.FileMetadata
		if err := rows.Scan(&meta.ID, &meta.FileID, &meta.MetadataKey, &meta.MetadataValue); err != nil {
			return nil, nil, status.Errorf(codes.Internal, "failed to scan file metadata: %v", err)
		}
	}

	return &file, fileMeta, nil
}
