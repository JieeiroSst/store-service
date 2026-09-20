package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewCouponRepository),
	fx.Provide(NewCouponRestrictionRepository),
	fx.Provide(NewCouponUsageRepository),
	fx.Provide(NewUserCouponRepository),
)
