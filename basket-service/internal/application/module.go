package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewBasketService),
	fx.Provide(NewBasketLineService),
	fx.Provide(NewBasketLineAttributeService),
	fx.Provide(NewOrderService),
	fx.Provide(NewUserService),
)
