package infrastructure

import (
	"github.com/JIeeiroSst/networking-service/config"
	dnsadapter "github.com/JIeeiroSst/networking-service/internal/adapter/primary/dns"
	httpadapter "github.com/JIeeiroSst/networking-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/probe"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/snapshot"
	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.FromEnv),

	fx.Provide(provideStore),
	probe.Module,
	snapshot.Module,
	metrics.Module,

	application.Module,

	httpadapter.Module,
	dnsadapter.Module,

	fx.Invoke(server.New),
)
