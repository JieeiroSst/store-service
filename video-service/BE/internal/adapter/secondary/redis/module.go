package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/video-service/internal/domain/port"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewClient,
		fx.Annotate(NewStore,
			fx.As(new(port.ViewCounter)),
			fx.As(new(port.JobQueue)),
			fx.As(new(port.InvalidationBus)),
		),
	),
	fx.Invoke(func(lc fx.Lifecycle, rdb *goredis.Client) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
				if err := rdb.Ping(ctx).Err(); err != nil {
					return fmt.Errorf("redis: %w", err)
				}
				return nil
			},
			OnStop: func(context.Context) error { return rdb.Close() },
		})
	}),
)
