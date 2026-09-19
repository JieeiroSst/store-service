package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewKitchenRepository),
	fx.Provide(NewFoodRepository),
	fx.Provide(NewCategoryRepository),
)
