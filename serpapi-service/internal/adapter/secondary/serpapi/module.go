package serpapi

import (
	"github.com/JIeeiroSst/serpapi-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(func(c *Client) port.SerpAPI { return c }),
)
