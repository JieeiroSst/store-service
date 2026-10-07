package application

import (
	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/domain"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() *domain.Catalog { return domain.NewCatalog(domain.Programs) }),
	fx.Provide(func(cfg *config.Config) Settings { return Settings{LimitsLocation: cfg.Limits.Location} }),
	fx.Provide(NewAccountService),
	fx.Provide(NewCardService),
	fx.Provide(NewAuthorizationService),
)
