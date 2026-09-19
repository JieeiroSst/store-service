package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/gateway-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/gateway-service/internal/adapter/secondary/proxy"
	"github.com/JIeeiroSst/gateway-service/internal/application"
	"github.com/JIeeiroSst/gateway-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	proxy.Module, // port.{Consumer,Accounting,Delivery}Proxy

	application.Module, // port.GatewayUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
