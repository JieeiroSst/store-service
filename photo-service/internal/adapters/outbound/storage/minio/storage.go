package minio

import (
	"context"
	"io"
	"time"

	miniogo "github.com/minio/minio-go/v7"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/pkg/config"
)

type Storage struct {
	client *miniogo.Client
	bucket string
}

func NewStorage(client *miniogo.Client, cfg *config.Config) *Storage {
	return &Storage{client: client, bucket: cfg.MinIO.Bucket}
}

var _ ports.ImageStorage = (*Storage)(nil)

func (s *Storage) Put(ctx context.Context, key string, r io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, -1, miniogo.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, miniogo.GetObjectOptions{})
}

func (s *Storage) PresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
