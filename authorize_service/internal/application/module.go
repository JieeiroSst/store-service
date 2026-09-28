package application

import (
	"context"

	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewEnforcerProvider),
	fx.Provide(NewCasbinService),
	fx.Provide(NewPermissionService),
	fx.Provide(NewOTPService),
	fx.Invoke(seedRoleCatalog),
)

func seedRoleCatalog(cfg *config.Config, perms port.PermissionUsecase) error {
	return perms.SeedDefaults(context.Background(), cfg.Bootstrap.SuperAdminUserIDs)
}
