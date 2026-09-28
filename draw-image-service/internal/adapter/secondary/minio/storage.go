package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/JIeeiroSst/draw-image-service/internal/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

type storage struct {
	client *minio.Client
	bucket string
}

func NewStorage(lc fx.Lifecycle, cfg *config.Config) (port.ObjectStorage, error) {
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	s := &storage{client: client, bucket: cfg.Minio.Bucket}

	lc.Append(fx.Hook{OnStart: s.ensureBucket})
	return s, nil
}

func (s *storage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket %s: %w", s.bucket, err)
	}
	return nil
}

func (s *storage) Bucket() string { return s.bucket }

func (s *storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err := obj.Stat(); err != nil {
		obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return obj, nil
}

func (s *storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
