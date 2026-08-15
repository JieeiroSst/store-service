package distribution

import "go.uber.org/fx"

var Module = fx.Module("distribution-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(DistributionService)))),
)
