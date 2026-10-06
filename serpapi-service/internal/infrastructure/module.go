package infrastructure

import (
	"github.com/JIeeiroSst/serpapi-service/config"
	httpadapter "github.com/JIeeiroSst/serpapi-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/cache"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/serpapi"
	"github.com/JIeeiroSst/serpapi-service/internal/application"
	"github.com/JIeeiroSst/serpapi-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.Load),

	serpapi.Module,
	cache.Module,
	metrics.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
