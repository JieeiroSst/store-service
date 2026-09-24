package model

import (
	"errors"
	"fmt"
	"time"
)

type File struct {
	ID          string    `json:"id" bson:"_id"`
	ReceiverID  string    `json:"receiver_id" bson:"receiver_id"`
	FileName    string    `json:"file_name" bson:"file_name"`
	ContentType string    `json:"content_type" bson:"content_type"`
	Size        int64     `json:"size" bson:"size"`
	SHA256      string    `json:"sha256" bson:"sha256"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	ObjectKey   string    `json:"-" bson:"object_key"`
	URL         string    `json:"url,omitempty" bson:"-"`
}

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalid         = errors.New("invalid argument")
	ErrTooLarge        = errors.New("file too large")
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUpstream        = errors.New("upstream unavailable")
)

func Invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}
