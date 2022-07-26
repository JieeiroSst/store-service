package model

import (
	"time"

	"github.com/qor/media/oss"
)

type Media struct {
	ID                   string
	URL                  string
	Thumbnail            oss.OSS
	Description          string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}