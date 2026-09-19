package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewParkingService),
	fx.Provide(NewRateService),
)
