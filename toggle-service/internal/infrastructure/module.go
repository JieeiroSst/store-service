package infrastructure

import (
	"go.uber.org/fx"

	httpadapter "github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/toggle-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/toggle-service/internal/application"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/logger"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/server"
)

var Module = fx.Options(
	config.Module,
	logger.Module,
	database.Module,
	repository.Module,
	application.Module,
	httpadapter.Module,

	fx.Invoke(database.RunMigrations),
	fx.Invoke(server.NewHTTPServer),
)
