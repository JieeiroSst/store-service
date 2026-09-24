package infrastructure

import (
	"log/slog"
	"os"

	"github.com/JIeeiroSst/catalogues-service/config"
	"github.com/JIeeiroSst/catalogues-service/internal/adapter/primary/httpapi"
	"github.com/JIeeiroSst/catalogues-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/catalogues-service/internal/application"
	"github.com/JIeeiroSst/catalogues-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/catalogues-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

func initLogger() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

// Core is everything except the database, so tests can swap the connection.
var Core = fx.Options(
	repository.Module,
	application.Module,
	httpapi.Module,
	fx.Invoke(server.New),
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(config.Load),
	database.Module,
	Core,
)
