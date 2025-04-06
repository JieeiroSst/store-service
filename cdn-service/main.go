package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"
)

type FileService struct {
	pb.UnimplementedFileServiceServer
	db      *sql.DB
	baseURL string
	baseDir string
}

func NewFileService(db *sql.DB, baseURL, baseDir string) *FileService {
	return &FileService{
		db:      db,
		baseURL: baseURL,
		baseDir: baseDir,
	}
}

func (s *FileService) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.FileResponse, error) {
	fileID := uuid.New().String()

	if req.FileType != pb.FileType_FILE_TYPE_IMAGE && req.FileType != pb.FileType_FILE_TYPE_VIDEO {
		return nil, status.Error(codes.InvalidArgument, "file type must be either IMAGE or VIDEO")
	}

	fileTypeStr := "images"
	if req.FileType == pb.FileType_FILE_TYPE_VIDEO {
		fileTypeStr = "videos"
	}

	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	storagePath := filepath.Join(fileTypeStr, datePath, fileID+filepath.Ext(req.Filename))

	fullPath := filepath.Join(s.baseDir, storagePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create directory: %v", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	fileTypeDB := "IMAGE"
	if req.FileType == pb.FileType_FILE_TYPE_VIDEO {
		fileTypeDB = "VIDEO"
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO files (id, filename, file_type, mime_type, size_bytes, storage_path, content_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, fileID, req.Filename, fileTypeDB, req.MimeType, len(req.Content), storagePath, "hash-placeholder")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save file metadata: %v", err)
	}

	if len(req.Metadata) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO file_metadata (id, file_id, metadata_key, metadata_value)
			VALUES ($1, $2, $3, $4)
		`)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to prepare metadata statement: %v", err)
		}
		defer stmt.Close()

		for key, value := range req.Metadata {
			metadataID := uuid.New().String()
			if _, err := stmt.ExecContext(ctx, metadataID, fileID, key, value); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to save metadata: %v", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	url := fmt.Sprintf("%s/v1/files/%s/content", s.baseURL, fileID)

	return &pb.FileResponse{
		FileId:    fileID,
		Filename:  req.Filename,
		SizeBytes: int64(len(req.Content)),
		MimeType:  req.MimeType,
		FileType:  req.FileType,
		Url:       url,
		CreatedAt: timestamppb.New(now),
		Metadata:  req.Metadata,
	}, nil
}

func (s *FileService) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.FileResponse, error) {
	var (
		fileID    string
		filename  string
		fileType  string
		mimeType  string
		sizeBytes int64
		path      string
		createdAt time.Time
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT id, filename, file_type, mime_type, size_bytes, storage_path, created_at
		FROM files
		WHERE id = $1 AND is_deleted = false
	`, req.FileId).Scan(&fileID, &filename, &fileType, &mimeType, &sizeBytes, &path, &createdAt)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "file with ID %s not found", req.FileId)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query file: %v", err)
	}

	pbFileType := pb.FileType_FILE_TYPE_IMAGE
	if fileType == "VIDEO" {
		pbFileType = pb.FileType_FILE_TYPE_VIDEO
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT metadata_key, metadata_value
		FROM file_metadata
		WHERE file_id = $1
	`, fileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query metadata: %v", err)
	}
	defer rows.Close()

	metadata := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan metadata: %v", err)
		}
		metadata[key] = value
	}

	url := fmt.Sprintf("%s/v1/files/%s/content", s.baseURL, fileID)

	return &pb.FileResponse{
		FileId:    fileID,
		Filename:  filename,
		SizeBytes: sizeBytes,
		MimeType:  mimeType,
		FileType:  pbFileType,
		Url:       url,
		CreatedAt: timestamppb.New(createdAt),
		Metadata:  metadata,
	}, nil
}

func (s *FileService) GetFileContent(ctx context.Context, req *pb.GetFileRequest) (*pb.FileContentResponse, error) {
	var (
		filename  string
		mimeType  string
		path      string
		isDeleted bool
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT filename, mime_type, storage_path, is_deleted
		FROM files
		WHERE id = $1
	`, req.FileId).Scan(&filename, &mimeType, &path, &isDeleted)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "file with ID %s not found", req.FileId)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query file: %v", err)
	}

	if isDeleted {
		return nil, status.Errorf(codes.NotFound, "file with ID %s has been deleted", req.FileId)
	}

	return &pb.FileContentResponse{
		MimeType: mimeType,
		Filename: filename,
	}, nil
}

func (s *FileService) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	query := `
		SELECT id, filename, file_type, mime_type, size_bytes, storage_path, created_at
		FROM files
		WHERE is_deleted = false
	`
	args := []interface{}{}
	argPos := 1

	if req.FileType != pb.FileType_FILE_TYPE_UNSPECIFIED {
		fileType := "IMAGE"
		if req.FileType == pb.FileType_FILE_TYPE_VIDEO {
			fileType = "VIDEO"
		}
		query += fmt.Sprintf(" AND file_type = $%d", argPos)
		args = append(args, fileType)
		argPos++
	}

	if req.CreatedAfter != nil {
		query += fmt.Sprintf(" AND created_at > $%d", argPos)
		args = append(args, req.CreatedAfter.AsTime())
		argPos++
	}

	if req.CreatedBefore != nil {
		query += fmt.Sprintf(" AND created_at < $%d", argPos)
		args = append(args, req.CreatedBefore.AsTime())
		argPos++
	}

	if req.PageToken != "" {
		query += fmt.Sprintf(" AND id > $%d", argPos)
		args = append(args, req.PageToken)
		argPos++
	} 

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argPos)
	args = append(args, pageSize+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query files: %v", err)
	}
	defer rows.Close()

	files := make([]*pb.FileResponse, 0, pageSize)
	var nextPageToken string

	for rows.Next() {
		if len(files) >= int(pageSize) {
			var id string
			if err := rows.Scan(&id, nil, nil, nil, nil, nil, nil); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to scan file ID: %v", err)
			}
			nextPageToken = id
			break
		}

		var (
			fileID    string
			filename  string
			fileType  string
			mimeType  string
			sizeBytes int64
			path      string
			createdAt time.Time
		)

		if err := rows.Scan(&fileID, &filename, &fileType, &mimeType, &sizeBytes, &path, &createdAt); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan file: %v", err)
		}

		pbFileType := pb.FileType_FILE_TYPE_IMAGE
		if fileType == "VIDEO" {
			pbFileType = pb.FileType_FILE_TYPE_VIDEO
		}

		url := fmt.Sprintf("%s/v1/files/%s/content", s.baseURL, fileID)

		files = append(files, &pb.FileResponse{
			FileId:    fileID,
			Filename:  filename,
			SizeBytes: sizeBytes,
			MimeType:  mimeType,
			FileType:  pbFileType,
			Url:       url,
			CreatedAt: timestamppb.New(createdAt),
		})
	}

	var totalCount int32
	countQuery := `
		SELECT COUNT(*)
		FROM files
		WHERE is_deleted = false
	`
	countArgs := make([]interface{}, len(args)-1)
	copy(countArgs, args[:len(args)-1])

	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count files: %v", err)
	}

	return &pb.ListFilesResponse{
		Files:         files,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*emptypb.Empty, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var (
		path      string
		isDeleted bool
	)

	err = tx.QueryRowContext(ctx, `
		SELECT storage_path, is_deleted
		FROM files
		WHERE id = $1
	`, req.FileId).Scan(&path, &isDeleted)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "file with ID %s not found", req.FileId)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query file: %v", err)
	}

	if isDeleted {
		return &emptypb.Empty{}, nil
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE files
		SET is_deleted = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, req.FileId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark file as deleted: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &emptypb.Empty{}, nil
}
