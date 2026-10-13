package http

import (
	"github.com/JIeeiroSst/gateway-service/pkg/authn"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(authn.FromEnv, NewHandler),
)
