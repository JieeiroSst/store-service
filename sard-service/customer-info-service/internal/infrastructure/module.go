package infrastructure

import (
	"github.com/JIeeiroSst/customer-info-service/config"
	httpadapter "github.com/JIeeiroSst/customer-info-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/clock"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/crypto"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/ekyc"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/storage"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/customer-info-service/internal/application"
	"github.com/JIeeiroSst/customer-info-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.Load),

	storage.Module,
	userservice.Module,
	ekyc.Module,
	crypto.Module,
	clock.Module,
	metrics.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
