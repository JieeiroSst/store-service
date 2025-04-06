package model

import (
	"time"
)

type File struct {
	ID          string    `json:"id" db:"id"`
	Filename    string    `json:"filename" db:"filename"`
	FileType    string    `json:"file_type" db:"file_type"` // Should be "IMAGE" or "VIDEO"
	MimeType    string    `json:"mime_type" db:"mime_type"`
	SizeBytes   int64     `json:"size_bytes" db:"size_bytes"`
	StoragePath string    `json:"storage_path" db:"storage_path"`
	ContentHash string    `json:"content_hash" db:"content_hash"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsDeleted   bool      `json:"is_deleted" db:"is_deleted"`
}

type FileMetadata struct {
    ID           string    `json:"id" db:"id"`
    FileID       string    `json:"file_id" db:"file_id"` // Foreign key referencing files table
    MetadataKey  string    `json:"metadata_key" db:"metadata_key"`
    MetadataValue string   `json:"metadata_value" db:"metadata_value"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}