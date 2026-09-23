package infrastructure

import (
	"github.com/JIeeiroSst/video-service/config"
	httpadapter "github.com/JIeeiroSst/video-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/video-service/internal/adapter/secondary/buffer"
	"github.com/JIeeiroSst/video-service/internal/adapter/secondary/ffmpeg"
	"github.com/JIeeiroSst/video-service/internal/adapter/secondary/minio"
	"github.com/JIeeiroSst/video-service/internal/adapter/secondary/redis"
	"github.com/JIeeiroSst/video-service/internal/application"
	"github.com/JIeeiroSst/video-service/internal/infrastructure/server"
	"github.com/JIeeiroSst/video-service/internal/infrastructure/worker"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var common = fx.Options(
	fx.Provide(func() *config.Config {
		_ = godotenv.Load()
		return config.Load()
	}),

	minio.Module,
	redis.Module,
	buffer.Module,
)

var APIModule = fx.Options(
	common,
	application.APIModule,
	httpadapter.Module,
	fx.Invoke(server.New),
)

// WorkerModule consumes transcode jobs (ROLE=worker).
var WorkerModule = fx.Options(
	common,
	ffmpeg.Module,
	application.WorkerModule,
	fx.Invoke(worker.New),
)
