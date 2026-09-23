package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	miniogo "github.com/minio/minio-go/v7"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewClient,
		fx.Annotate(NewStorage, fx.As(new(port.VideoStorage))),
		fx.Annotate(NewRepository, fx.As(new(port.VideoRepository))),
	),
	fx.Invoke(ensureBucket),
)

func ensureBucket(lc fx.Lifecycle, client *miniogo.Client, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			exists, err := client.BucketExists(ctx, cfg.Minio.Bucket)
			if err != nil {
				return fmt.Errorf("check bucket %q: %w", cfg.Minio.Bucket, err)
			}
			if exists {
				return nil
			}
			if err := client.MakeBucket(ctx, cfg.Minio.Bucket, miniogo.MakeBucketOptions{}); err != nil {
				return fmt.Errorf("create bucket %q: %w", cfg.Minio.Bucket, err)
			}
			return nil
		},
	})
}
