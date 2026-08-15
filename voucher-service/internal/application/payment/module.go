package payment

import "go.uber.org/fx"

var Module = fx.Module("payment-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(PaymentService)))),
)
