package objectstore

import (
	"context"
	"fmt"
	"io"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

type Store struct {
	client *minio.Client
	bucket string
}

func New(cfg *config.Config) (*Store, error) {
	s := cfg.Storage
	cl, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.SecretKey, ""),
		Secure: s.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	return &Store{client: cl, bucket: s.Bucket}, nil
}

func (s *Store) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if exists {
		return nil
	}
	err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	if err != nil && minio.ToErrorResponse(err).Code != "BucketAlreadyOwnedByYou" {
		return fmt.Errorf("make bucket: %w", err)
	}
	return nil
}

func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("%w: object store: %v", model.ErrUpstream, err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: object store: %v", model.ErrUpstream, err)
	}

	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("%w: object store: %v", model.ErrUpstream, err)
	}
	return obj, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("%w: object store: %v", model.ErrUpstream, err)
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(New, func(s *Store) port.ObjectStore { return s }),
	fx.Invoke(func(lc fx.Lifecycle, s *Store) {
		lc.Append(fx.Hook{OnStart: s.EnsureBucket})
	}),
)
