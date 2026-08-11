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

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	redis.Module,    // *redis.Client

	repository.Module, // port.RoomRepository, port.StreamRepository, port.VODRepository
	redisstore.Module, // port.NodeRegistry, port.ViewerCounter, port.ChatBroadcaster
	storage.Module,    // port.ObjectStorage
	transcode.Module,  // port.TranscodeRunner

	application.Module, // port.RoomUsecase, port.NodeSchedulerUsecase, port.StreamLifecycleUsecase, port.ViewerUsecase, port.ChatUsecase

	httpadapter.Module,      // *httpadapter.Handler, *httpadapter.SRSWebhookHandler, *httpadapter.WSHandler
	schedulerAdapter.Module, // *scheduler.Heartbeat, *scheduler.Watchdog

	fx.Invoke(server.NewHTTPServer),
	fx.Invoke(server.NewSchedulerServer),
)
