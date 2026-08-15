package auth

import "go.uber.org/fx"

var Module = fx.Module("auth-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(AuthService)))),
)
