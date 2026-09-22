package app

import (
	"github.com/JIeeiroSst/partner-service/internal/adapters/handler"
	"github.com/JIeeiroSst/partner-service/internal/adapters/repository"
	"github.com/JIeeiroSst/partner-service/internal/adapters/router"
	"github.com/JIeeiroSst/partner-service/internal/adapters/scheduler"
	"github.com/JIeeiroSst/partner-service/internal/adapters/server"
	"github.com/JIeeiroSst/partner-service/internal/config"
	"github.com/JIeeiroSst/partner-service/internal/consul"
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/JIeeiroSst/partner-service/internal/logger"
	"go.uber.org/fx"
)

// newConfig lives here, not in the config package, because it needs the
// consul package to fetch the config and consul already imports config —
// putting it in either of those would create an import cycle.
func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	return consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
}

func initLogger() {
	logger.SetupLogger()
}

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	repository.Module, // ports.PartnerRepository, ports.PartnershipRepository, ports.PartnershipsPartnerRepository, ports.ProjectRepository
	services.Module,   // *services.PartnerService, *services.PartnershipService, *services.PartnershipsPartnerService, *services.ProjectService
	handler.Module,    // *handler.PartnerHandler, *handler.PartnershipHandler, *handler.PartnershipsPartnerHandler, *handler.ProjectHandler

	fx.Provide(router.NewRouter), // *gin.Engine

	scheduler.Module, // daily inactive-partner sweep, started/stopped with the app

	fx.Invoke(server.New),
)
