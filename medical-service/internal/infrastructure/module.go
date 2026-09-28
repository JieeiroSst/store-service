package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/medical-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/medical-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/medical-service/internal/application"
	"github.com/JIeeiroSst/medical-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/medical-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),
	fx.Provide(newClock),

	database.Module,

	repository.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
