package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewKitchenService),
	fx.Provide(NewFoodService),
	fx.Provide(NewCategoryService),
)
