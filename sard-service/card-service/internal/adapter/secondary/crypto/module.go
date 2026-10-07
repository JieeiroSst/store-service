package crypto

import (
	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) Keys { return DeriveKeys([]byte(cfg.Security.MasterKey)) }),
	fx.Provide(func(k Keys) (port.Vault, error) { return NewVault(k) }),
	fx.Provide(func(k Keys) port.CardSecurity { return NewSecurity(k) }),
)
