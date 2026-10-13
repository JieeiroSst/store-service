package infrastructure

import (
	"log/slog"
	"os"
	"strings"
	"time"

	httpadapter "github.com/JIeeiroSst/wishlists-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/wishlists-service/internal/adapter/secondary/postgres"
	rediscache "github.com/JIeeiroSst/wishlists-service/internal/adapter/secondary/redis"
	"github.com/JIeeiroSst/wishlists-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/wishlists-service/internal/application"
	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/JIeeiroSst/wishlists-service/internal/infrastructure/cache"
	"github.com/JIeeiroSst/wishlists-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/wishlists-service/internal/infrastructure/server"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Module = fx.Options(
	fx.Provide(config.Load),
	fx.Invoke(setupLogging),

	database.Module,
	cache.Module,
	postgres.Module,
	userservice.Module,
	rediscache.Module,
	application.Module,
	httpadapter.Module,

	fx.Invoke(server.New),
)

func Options() []fx.Option {
	return []fx.Option{
		Module,
		fx.WithLogger(func() fxevent.Logger {
			l := &fxevent.SlogLogger{Logger: slog.Default().With("component", "fx")}
			l.UseLogLevel(slog.LevelDebug)
			return l
		}),
		fx.StopTimeout(20 * time.Second),
	}
}

func setupLogging(cfg *config.Config) {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
