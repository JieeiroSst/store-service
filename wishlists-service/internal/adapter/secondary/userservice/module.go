package userservice

import (
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.UserDirectory)))),
)
