package infrastructure

import (
	"github.com/JIeeiroSst/webrtc-service/config"
	httpadapter "github.com/JIeeiroSst/webrtc-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/webrtc-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/webrtc-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/webrtc-service/internal/application"
	"github.com/JIeeiroSst/webrtc-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.FromEnv),

	memory.Module,
	metrics.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
