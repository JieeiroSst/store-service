package model

import "time"

type FileType string

const (
	FileTypeImage FileType = "IMAGE"
	FileTypeVideo FileType = "VIDEO"
)

type File struct {
	ID          string
	Filename    string
	FileType    FileType
	MimeType    string
	SizeBytes   int64
	StoragePath string
	ContentHash string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

type FileMetadata struct {
	ID            string
	FileID        string
	MetadataKey   string
	MetadataValue string
	CreatedAt     time.Time
}

type UploadFileInput struct {
	Content  []byte
	Filename string
	MimeType string
	FileType FileType
	Metadata map[string]string
}

type FileResponse struct {
	FileID      string
	Filename    string
	SizeBytes   int64
	MimeType    string
	FileType    FileType
	URL         string
	CreatedAt   time.Time
	Metadata    map[string]string
	StoragePath string
}
