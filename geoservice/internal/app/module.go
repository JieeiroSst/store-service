package app

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/geoservice/config"
	"github.com/geoservice/internal/adapter/handler"
	"github.com/geoservice/internal/adapter/migration"
	"github.com/geoservice/internal/adapter/repository"
	"github.com/geoservice/internal/core/service"
)

var Module = fx.Options(
	fx.Provide(
		config.Load,
		NewLogger,
		NewPool,
		repository.NewGeoRepository,
		service.NewLocatorService,
	),
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.LogLevel == "debug" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func NewPool(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return err
			}
			log.Info("database connected")
			return migration.Run(ctx, pool, log)
		},
		OnStop: func(context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil
}

var ServerModule = fx.Options(
	Module,
	fx.Provide(handler.NewLocationHandler),
	fx.Invoke(RunHTTPServer),
)

func RunHTTPServer(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger, h *handler.LocationHandler) {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	h.Register(e)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				addr := ":" + cfg.HTTPPort
				log.Info("http server listening", zap.String("addr", addr))
				if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
					log.Error("server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})
}
