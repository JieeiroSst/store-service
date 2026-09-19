package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewVehicleRepository),
	fx.Provide(NewParkingSpotRepository),
	fx.Provide(NewTicketRepository),
	fx.Provide(NewRateRepository),
	fx.Provide(NewPaymentRepository),
)
