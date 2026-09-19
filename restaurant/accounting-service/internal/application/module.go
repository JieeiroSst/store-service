package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAuthCartService),
	fx.Provide(NewPaymentService),
)
