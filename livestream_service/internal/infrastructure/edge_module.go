package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/livestream-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/redisstore"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/livestream-service/internal/application"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/redis"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var EdgeModule = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module,
	redis.Module,

	repository.Module,
	redisstore.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.NewEdgeHTTPServer),
)
