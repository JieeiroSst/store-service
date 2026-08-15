package http

import "go.uber.org/fx"

var Module = fx.Module("http",
	fx.Provide(NewCompositionHandler),
	fx.Provide(NewRouter),
	fx.Invoke(RegisterServer),
)
