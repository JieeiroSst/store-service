package grpcapi

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewHospitalHandler, NewAppointmentHandler),
)
