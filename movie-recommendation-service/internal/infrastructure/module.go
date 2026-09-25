package infrastructure

import (
	"github.com/JIeeiroSst/movie-recommendation-service/config"
	httpadapter "github.com/JIeeiroSst/movie-recommendation-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/adapter/secondary/postgres"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/adapter/secondary/redis"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/adapter/secondary/videoclient"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/application"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/infrastructure/jobs"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/infrastructure/server"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() *config.Config {
		_ = godotenv.Load()
		return config.Load()
	}),

	postgres.Module,
	fx.Provide(newSnapshotRepository),
	videoclient.Module,
	application.Module,
	httpadapter.Module,

	fx.Invoke(server.New),
	fx.Invoke(jobs.New),
)

func newSnapshotRepository(lc fx.Lifecycle, cfg *config.Config, pg *postgres.SnapshotStore) port.SnapshotRepository {
	if cfg.Redis.Addr != "" {
		return redis.New(lc, cfg)
	}
	return pg
}
