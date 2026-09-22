package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewCitizenIdentityRepository),
	fx.Provide(NewFaceBiometricRepository),
	fx.Provide(NewVerificationRepository),
)
