package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewProductRepository),
	fx.Provide(NewOrderRepository),
	fx.Provide(NewWebhookEventRepository),
)
