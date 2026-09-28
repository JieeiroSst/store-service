package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewShipmentService),
	fx.Provide(NewWebhookService),
	fx.Provide(NewWarehouseService),
	fx.Provide(NewLocationService),
	fx.Provide(NewOutboxService),
)
