package application

import (
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() *domain.Catalog { return domain.NewCatalog(domain.Engines) }),
	fx.Provide(NewLimiter),
	fx.Provide(NewSearchService),
	fx.Provide(NewAccountService),
	fx.Provide(NewImageService),
)
