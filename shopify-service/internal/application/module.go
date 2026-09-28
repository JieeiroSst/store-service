package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewProductService),
	fx.Provide(NewOrderService),
	fx.Provide(NewWebhookService),
)
