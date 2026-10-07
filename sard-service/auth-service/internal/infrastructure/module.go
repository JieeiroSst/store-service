package infrastructure

import (
	"github.com/JIeeiroSst/auth-service/config"
	httpadapter "github.com/JIeeiroSst/auth-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/authorizeservice"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/cardservice"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/clock"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/otp"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/storage"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/auth-service/internal/application"
	"github.com/JIeeiroSst/auth-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.Load),

	storage.Module,
	cardservice.Module,
	userservice.Module,
	authorizeservice.Module,
	otp.Module,
	clock.Module,
	metrics.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
