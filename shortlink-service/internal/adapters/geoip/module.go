package geoip

import (
	"context"

	"github.com/JIeeiroSst/shortlink-service/internal/config"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("geoip",
	fx.Provide(NewLookupFromConfig),
)

func NewLookupFromConfig(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (ports.GeoIPLookup, error) {
	lookup, err := New(cfg.GeoIPDBPath, log)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return lookup.Close()
		},
	})
	return lookup, nil
}
