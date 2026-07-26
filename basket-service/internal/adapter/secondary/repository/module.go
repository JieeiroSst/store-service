package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewBasketRepository),
	fx.Provide(NewBasketLineRepository),
	fx.Provide(NewBasketLineAttributeRepository),
	fx.Provide(NewOrderRepository),
	fx.Provide(NewUserRepository),
)
