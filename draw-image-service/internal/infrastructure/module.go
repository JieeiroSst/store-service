package infrastructure

import (
	"github.com/JIeeiroSst/draw-image-service/config"
	httpadapter "github.com/JIeeiroSst/draw-image-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/draw-image-service/internal/adapter/secondary/imaging"
	"github.com/JIeeiroSst/draw-image-service/internal/adapter/secondary/minio"
	"github.com/JIeeiroSst/draw-image-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/draw-image-service/internal/application"
	"github.com/JIeeiroSst/draw-image-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/draw-image-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.FromEnv),

	database.Module,
	repository.Module,
	minio.Module,
	imaging.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
