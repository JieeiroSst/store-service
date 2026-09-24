package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

const maxObject = 64 << 20

type store struct {
	client *minio.Client
	bucket string

	mu     sync.Mutex
	ensure bool
}

func New(lc fx.Lifecycle, cfg *config.Config) (port.DocumentStore, error) {
	sc := cfg.Storage
	if sc.Endpoint == "" {
		return disabled{}, nil
	}
	cl, err := minio.New(sc.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(sc.AccessKey, sc.SecretKey, ""),
		Secure: sc.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	s := &store{client: cl, bucket: sc.Bucket}

	lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		if err := s.ensureBucket(ctx); err != nil {
			logrus.WithError(err).Warn("document bucket not ready yet")
		}
		return nil
	}})
	return s, nil
}

func (s *store) Enabled() bool { return true }

func (s *store) ensureBucket(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ensure {
		return nil
	}
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			if minio.ToErrorResponse(err).Code != "BucketAlreadyOwnedByYou" {
				return err
			}
		}
	}
	s.ensure = true
	return nil
}

func (s *store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	put := func() error {
		_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)),
			minio.PutObjectOptions{ContentType: contentType})
		return err
	}
	err := put()
	if minio.ToErrorResponse(err).Code == "NoSuchBucket" {
		if err = s.ensureBucket(ctx); err == nil {
			err = put()
		}
	}
	if err != nil {
		return fmt.Errorf("%w: document store: %v", common.ErrUpstream, err)
	}
	return nil
}

func (s *store) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: document store: %v", common.ErrUpstream, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(io.LimitReader(obj, maxObject+1))
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("%w: document store: %v", common.ErrUpstream, err)
	}
	if len(data) > maxObject {
		return nil, fmt.Errorf("%w: document store: object too large", common.ErrUpstream)
	}
	return data, nil
}

func (s *store) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("%w: document store: %v", common.ErrUpstream, err)
	}
	return nil
}

type disabled struct{}

func (disabled) Enabled() bool { return false }
func (disabled) Put(context.Context, string, []byte, string) error {
	return fmt.Errorf("%w: document store is not configured", common.ErrUpstream)
}
func (disabled) Get(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("%w: document store is not configured", common.ErrUpstream)
}
func (disabled) Delete(context.Context, string) error { return nil }

var Module = fx.Options(fx.Provide(New))
