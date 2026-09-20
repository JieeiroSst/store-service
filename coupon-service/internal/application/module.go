package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewCouponService),
	fx.Provide(NewRestrictionService),
	fx.Provide(NewUsageService),
	fx.Provide(NewUserCouponService),
)
