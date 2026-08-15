package reporting

import "go.uber.org/fx"

var Module = fx.Module("reporting-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(ReportingService)))),
)
