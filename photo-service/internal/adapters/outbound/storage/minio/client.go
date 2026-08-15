package minio

import (
	"context"
	"fmt"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/photo-service/pkg/config"
)

func NewClient(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*miniogo.Client, error) {
	client, err := miniogo.New(cfg.MinIO.Endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			exists, err := client.BucketExists(ctx, cfg.MinIO.Bucket)
			if err != nil {
				return fmt.Errorf("check bucket %q: %w", cfg.MinIO.Bucket, err)
			}
			if !exists {
				if err := client.MakeBucket(ctx, cfg.MinIO.Bucket, miniogo.MakeBucketOptions{}); err != nil {
					return fmt.Errorf("create bucket %q: %w", cfg.MinIO.Bucket, err)
				}
				log.Info("created minio bucket", zap.String("bucket", cfg.MinIO.Bucket))
			} else {
				log.Info("minio bucket already exists", zap.String("bucket", cfg.MinIO.Bucket))
			}
			return nil
		},
	})

	return client, nil
}
