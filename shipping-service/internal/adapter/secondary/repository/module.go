package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewTxManager),
	fx.Provide(NewShipmentRepository),
	fx.Provide(NewEventRepository),
	fx.Provide(NewWarehouseRepository),
	fx.Provide(NewOutboxRepository),
)
