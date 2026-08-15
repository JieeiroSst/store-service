package imaging

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
)

var Module = fx.Module("imaging",
	fx.Provide(NewComposer),
	fx.Provide(func(c *Composer) ports.ImageComposer { return c }),
	fx.Provide(NewDecoder),
	fx.Provide(func(d *Decoder) ports.ImageDecoder { return d }),
	fx.Provide(NewFetcher),
	fx.Provide(func(f *Fetcher) ports.ImageFetcher { return f }),
)
