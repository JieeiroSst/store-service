package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type objectStorage struct {
	client *minio.Client
	bucket string
}

func NewObjectStorage(cfg *config.Config) (port.ObjectStorage, error) {
	client, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	s := &objectStorage{client: client, bucket: cfg.Storage.Bucket}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, s.bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket %q: %w", s.bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %q: %w", s.bucket, err)
		}
	}
	return s, nil
}

func (s *objectStorage) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType, cacheControl string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: cacheControl,
	})
	return err
}

func (s *objectStorage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
