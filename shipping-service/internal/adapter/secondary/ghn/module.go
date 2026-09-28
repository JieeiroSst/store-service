package ghn

import (
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(fx.Annotate(func(c *Client) *Client { return c }, fx.As(new(port.Carrier)))),
	fx.Provide(fx.Annotate(NewLocations, fx.As(new(port.LocationDirectory)))),
)
