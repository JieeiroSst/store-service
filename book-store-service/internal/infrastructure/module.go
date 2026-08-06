package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/bookStore-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/bookStore-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/bookStore-service/internal/application"
	"github.com/JIeeiroSst/bookStore-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/bookStore-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.*Repository

	application.Module, // port.*Usecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
