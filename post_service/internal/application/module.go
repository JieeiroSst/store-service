package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPostService),
	fx.Provide(NewCategoryService),
)
