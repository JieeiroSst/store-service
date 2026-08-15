package merchant

import "go.uber.org/fx"

var Module = fx.Module("merchant-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(MerchantService)))),
)
