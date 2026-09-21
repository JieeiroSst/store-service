package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewOrderRepository),
	fx.Provide(NewOrderTrackingRepository),
	fx.Provide(NewDriverAssignmentRepository),
)
