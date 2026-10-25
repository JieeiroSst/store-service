package application

import (
	"context"

	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/application/mfa"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewEnforcerProvider),
	fx.Provide(NewCasbinService),
	fx.Provide(NewPermissionService),
	fx.Provide(NewOTPService),
	fx.Provide(newMFAService),
	fx.Invoke(seedRoleCatalog),
)

func newMFAService(cfg *config.Config, repo port.TOTPRepository, box port.SecretBox, qr port.QRCodeEncoder, throttle port.FailureThrottle) port.MFAUsecase {
	return mfa.New(repo, box, qr, throttle, mfa.Config{Issuer: cfg.MFA.IssuerName()})
}

func seedRoleCatalog(cfg *config.Config, perms port.PermissionUsecase) error {
	return perms.SeedDefaults(context.Background(), cfg.Bootstrap.SuperAdminUserIDs)
}
