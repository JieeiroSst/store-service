package infrastructure

import (
	"github.com/JIeeiroSst/room-service/config"
	httpadapter "github.com/JIeeiroSst/room-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/room-service/internal/adapter/primary/ws"
	"github.com/JIeeiroSst/room-service/internal/adapter/secondary/eventbus"
	"github.com/JIeeiroSst/room-service/internal/adapter/secondary/mysql"
	"github.com/JIeeiroSst/room-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/room-service/internal/application"
	"github.com/JIeeiroSst/room-service/internal/infrastructure/server"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() *config.Config {
		_ = godotenv.Load()
		return config.Load()
	}),

	mysql.Module,
	userservice.Module,
	// The hub is the local half of the event bus.
	fx.Provide(func(h *ws.Hub) eventbus.Local { return h }),
	eventbus.Module,

	application.Module,
	httpadapter.Module,
	fx.Invoke(server.New),
)
