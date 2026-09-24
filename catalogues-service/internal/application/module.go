package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewCategoryService,
		NewProductClassService,
		NewOptionService,
		NewProductService,
	),
)
