package keycloak

import (
	"github.com/JIeeiroSst/manage-service/config"
	"github.com/Nerzal/gocloak/v13"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(NewIdentityProvider),
)

func NewClient(cfg *config.Config) *gocloak.GoCloak {
	if cfg.Keycloak.LegacyWildfly {
		return gocloak.NewClient(cfg.Keycloak.Host, gocloak.SetLegacyWildFlySupport())
	}
	return gocloak.NewClient(cfg.Keycloak.Host)
}
