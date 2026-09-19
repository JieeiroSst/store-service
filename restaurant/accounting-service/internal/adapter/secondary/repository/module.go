package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAuthCartRepository),
	fx.Provide(NewPaymentRepository),
)
