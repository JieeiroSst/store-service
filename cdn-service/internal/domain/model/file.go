package model

import (
	"time"

	"github.com/google/uuid"
)

type FileStatus string

const (
	FileStatusPending  FileStatus = "pending"
	FileStatusUploaded FileStatus = "uploaded"
	FileStatusFailed   FileStatus = "failed"
)

type File struct {
	ID          uuid.UUID
	ObjectKey   string
	FileName    string
	ContentType string
	SizeBytes   int64
	Status      FileStatus
	UploadedBy  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
