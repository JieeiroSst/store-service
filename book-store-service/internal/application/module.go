package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewBookService),
	fx.Provide(NewAuthorService),
	fx.Provide(NewPublisherService),
	fx.Provide(NewCategoryService),
)
