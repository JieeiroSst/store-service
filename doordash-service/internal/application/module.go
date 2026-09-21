package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewOrderService),
	fx.Provide(NewTrackingService),
	fx.Provide(NewDriverAssignmentService),
)
