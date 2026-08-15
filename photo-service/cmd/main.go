package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/photo-service/internal/adapters/inbound/http"
	"github.com/JIeeiroSst/photo-service/internal/adapters/outbound/cache/redis"
	"github.com/JIeeiroSst/photo-service/internal/adapters/outbound/imaging"
	"github.com/JIeeiroSst/photo-service/internal/adapters/outbound/persistence/pg"
	"github.com/JIeeiroSst/photo-service/internal/adapters/outbound/storage/minio"
	"github.com/JIeeiroSst/photo-service/internal/application"
	"github.com/JIeeiroSst/photo-service/pkg/config"
	"github.com/JIeeiroSst/photo-service/pkg/logger"
)

func main() {
	fx.New(
		config.Module,
		logger.Module,

		// Outbound
		pg.Module,
		redis.Module,
		minio.Module,
		imaging.Module,

		// Core use cases.
		application.Module,

		// Inbound (driving) adapter.
		http.Module,

		fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
			return logger.FxLogger(l)
		}),
	).Run()
}
