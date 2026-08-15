package partner

import (
	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"go.uber.org/fx"
)

func newService(repo APIKeyRepository, cfg *config.Config) PartnerService {
	return NewService(repo, cfg.PartnerHMACEncKey)
}

var Module = fx.Module("partner-app",
	fx.Provide(fx.Annotate(newService, fx.As(new(PartnerService)))),
)
