package publisher

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewNotificationPublisher),
	fx.Provide(NewBatchPublisher),
)
