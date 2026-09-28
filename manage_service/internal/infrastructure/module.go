package infrastructure

import (
	"time"

	httpadapter "github.com/JIeeiroSst/manage-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/manage-service/internal/adapter/secondary/keycloak"
	"github.com/JIeeiroSst/manage-service/internal/application"
	"github.com/JIeeiroSst/manage-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

const StopTimeout = 30 * time.Second

var Module = fx.Options(
	fx.Provide(newConfig),
	fx.Provide(newLogger),
	fx.Provide(newAdminCredentials),

	keycloak.Module,
	application.Module,
	httpadapter.Module,

	fx.Invoke(server.New),
)
