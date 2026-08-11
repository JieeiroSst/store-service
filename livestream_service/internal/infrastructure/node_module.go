package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/livestream-service/internal/adapter/primary/http"
	schedulerAdapter "github.com/JIeeiroSst/livestream-service/internal/adapter/primary/scheduler"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/redisstore"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/storage"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/transcode"
	"github.com/JIeeiroSst/livestream-service/internal/application"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/redis"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var NodeModule = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module,
	redis.Module,

	repository.Module,
	redisstore.Module,
	storage.Module,
	transcode.Module, // port.TranscodeRunner - node role only

	application.Module,

	httpadapter.Module,
	schedulerAdapter.Module, // *scheduler.Heartbeat, *scheduler.Watchdog - node role only

	fx.Invoke(server.NewNodeHTTPServer),
	fx.Invoke(server.NewSchedulerServer),
)
