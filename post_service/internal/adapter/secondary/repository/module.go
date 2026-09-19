package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPostRepository),
	fx.Provide(NewCategoryRepository),
	fx.Provide(NewMediaRepository),
)
