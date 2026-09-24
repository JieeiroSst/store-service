package httpapi

import "go.uber.org/fx"

func asHandler(constructor any) any {
	return fx.Annotate(constructor, fx.As(new(Handler)), fx.ResultTags(`group:"handlers"`))
}

var Module = fx.Options(
	fx.Provide(
		asHandler(NewCategoryHandler),
		asHandler(NewProductClassHandler),
		asHandler(NewOptionHandler),
		asHandler(NewProductHandler),
	),
)
