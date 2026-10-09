package infrastructure

import (
	"github.com/JIeeiroSst/search-service/config"
	httpadapter "github.com/JIeeiroSst/search-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/search-service/internal/adapter/secondary/elasticsearch"
	"github.com/JIeeiroSst/search-service/internal/application"
	"github.com/JIeeiroSst/search-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.Load),

	elasticsearch.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
