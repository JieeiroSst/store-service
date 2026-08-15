package order

import "go.uber.org/fx"

var Module = fx.Module("order-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(OrderService)))),
)
