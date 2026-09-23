package minio

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClient(cfg *config.Config) (*minio.Client, error) {
	return minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
}

type Storage struct {
	client *minio.Client
	bucket string
}

func NewStorage(client *minio.Client, cfg *config.Config) *Storage {
	return &Storage{client: client, bucket: cfg.Minio.Bucket}
}

func (s *Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (int64, error) {
	info, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (s *Storage) ReadAt(ctx context.Context, key string, size int64, p []byte, off int64) (int, error) {
	if off >= size {
		return 0, io.EOF
	}
	n := min(int64(len(p)), size-off)
	if n <= 0 {
		return 0, nil
	}
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(off, off+n-1); err != nil {
		return 0, err
	}
	obj, err := s.client.GetObject(ctx, s.bucket, key, opts)
	if err != nil {
		return 0, err
	}
	defer obj.Close()

	got, err := io.ReadFull(obj, p[:n])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return got, err
	}
	if int64(got) < n {
		return got, fmt.Errorf("short read of %s at %d: got %d of %d", key, off, got, n)
	}
	return got, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *Storage) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, notFoundOr(err)
	}
	return data, nil
}

func (s *Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Storage) DeletePrefix(ctx context.Context, prefix string) error {
	objs := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true})
	var firstErr error
	for e := range s.client.RemoveObjects(ctx, s.bucket, objs, minio.RemoveObjectsOptions{}) {
		if firstErr == nil {
			firstErr = e.Err
		}
	}
	return firstErr
}

func notFoundOr(err error) error {
	if minio.ToErrorResponse(err).Code == "NoSuchKey" {
		return port.ErrNotFound
	}
	return err
}
