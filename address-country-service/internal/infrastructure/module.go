package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/address-country-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/address-country-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/address-country-service/internal/application"
	"github.com/JIeeiroSst/address-country-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/address-country-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.*Repository implementations

	application.Module, // port.*Usecase implementations

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
