package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/shipping-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/shipping-service/internal/adapter/secondary/callback"
	"github.com/JIeeiroSst/shipping-service/internal/adapter/secondary/ghn"
	"github.com/JIeeiroSst/shipping-service/internal/adapter/secondary/notification"
	"github.com/JIeeiroSst/shipping-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/shipping-service/internal/application"
	"github.com/JIeeiroSst/shipping-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/shipping-service/internal/infrastructure/server"
	"github.com/JIeeiroSst/shipping-service/internal/infrastructure/worker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),
	fx.Provide(newClock),
	fx.Provide(newCodeGenerator),
	fx.Provide(newSettings),

	database.Module,
	repository.Module,

	ghn.Module,
	notification.Module,
	callback.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
	fx.Invoke(worker.RunOutbox),
)
