package domain

import (
	"bytes"
	"time"
)

const (
	MaxImageBytes = 500 << 10
	ImageIDTTL    = 10 * time.Minute
)

type ImageUpload struct {
	Filename    string
	ContentType string
	Data        []byte
}

type UploadedImage struct {
	ImageID   string    `json:"image_id"`
	Message   string    `json:"message,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewImageUpload(filename string, data []byte) (ImageUpload, error) {
	if len(data) == 0 {
		return ImageUpload{}, Invalid("image is empty")
	}
	if len(data) > MaxImageBytes {
		return ImageUpload{}, Invalid("image is %d bytes, the maximum is %d", len(data), MaxImageBytes)
	}
	ctype, ext := sniffImage(data)
	if ctype == "" {
		return ImageUpload{}, Invalid("image must be jpg, jpeg, png or webp")
	}
	if filename == "" {
		filename = "image." + ext
	}
	return ImageUpload{Filename: filename, ContentType: ctype, Data: data}, nil
}

func sniffImage(b []byte) (string, string) {
	switch {
	case bytes.HasPrefix(b, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", "jpg"
	case bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")):
		return "image/png", "png"
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return "image/webp", "webp"
	}
	return "", ""
}
