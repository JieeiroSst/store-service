package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		func(lc fx.Lifecycle, cfg *config.Config) (*pgxpool.Pool, error) {
			pool, err := NewPool(context.Background(), cfg)
			if err != nil {
				return nil, err
			}
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()
					if err := ensureDatabase(ctx, cfg); err != nil {
						return err
					}
					if err := Migrate(ctx, pool); err != nil {
						return fmt.Errorf("migrate: %w", err)
					}
					return nil
				},
				OnStop: func(context.Context) error { pool.Close(); return nil },
			})
			return pool, nil
		},
		fx.Annotate(NewRepository,
			fx.As(new(port.VideoRepository)),
			fx.As(new(port.InteractionRepository)),
			fx.As(new(port.HealthChecker)),
		),
		NewSnapshotStore,
	),
)
