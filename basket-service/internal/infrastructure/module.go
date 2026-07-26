package infrastructure

import (
	cronadapter "github.com/JIeeiroSst/basket-service/internal/adapter/primary/cron"
	httpadapter "github.com/JIeeiroSst/basket-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/basket-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/basket-service/internal/application"
	"github.com/JIeeiroSst/basket-service/internal/infrastructure/cache"
	"github.com/JIeeiroSst/basket-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/basket-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	cache.Module, // *redis.Client

	repository.Module, // port.*Repository

	application.Module, // port.*Usecase

	httpadapter.Module, // *httpadapter.Handler

	cronadapter.Module, // scheduled jobs

	fx.Invoke(server.New),
)
