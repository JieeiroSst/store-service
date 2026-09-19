package model

import "time"

// Media is an uploaded file's metadata (see
// internal/adapter/secondary/objectstorage for where the file itself
// lives). Thumbnail was previously a qor/media oss.OSS value that nothing
// ever populated or read - dropped in favor of a plain URL, which is all
// this service actually produces (see ObjectStorage.UploadFile's result).
type Media struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	URL         string    `json:"url" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Media) TableName() string { return "media" }
