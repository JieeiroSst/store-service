package proxy

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewConsumerProxy),
	fx.Provide(NewAccountingProxy),
	fx.Provide(NewDeliveryProxy),
)
