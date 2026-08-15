package corporate

import "go.uber.org/fx"

var Module = fx.Module("corporate-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(CorporateService)))),
)
