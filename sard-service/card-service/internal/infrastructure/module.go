package infrastructure

import (
	"github.com/JIeeiroSst/card-service/config"
	httpadapter "github.com/JIeeiroSst/card-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/authservice"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/clock"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/crypto"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/customerinfo"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/storage"
	"github.com/JIeeiroSst/card-service/internal/application"
	"github.com/JIeeiroSst/card-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.Load),

	storage.Module,
	customerinfo.Module,
	authservice.Module,
	crypto.Module,
	clock.Module,
	metrics.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
