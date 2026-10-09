package elasticsearch

import (
	"github.com/JIeeiroSst/search-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(NewSearcher),
	fx.Provide(func(s *Searcher) port.DocumentSearcher { return s }),
	fx.Provide(fx.Annotate(NewSchemaManager, fx.As(new(port.SchemaManager)))),
	fx.Provide(fx.Annotate(NewHealth, fx.As(new(port.HealthChecker)))),
)
