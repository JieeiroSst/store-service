package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
)

type ForecastProvider interface {
	GetForecast(ctx context.Context, coord model.Coordinate) (*model.Forecast, error)
}

type TideProvider interface {
	GetTidePredictions(ctx context.Context, stationID string) (*model.TidePrediction, error)
}

type RadarProvider interface {
	GetLatestRadar(ctx context.Context) (*model.RainRadar, error)
}

type WeatherCache interface {
	GetForecast(ctx context.Context, key string) (*model.Forecast, bool)
	SetForecast(ctx context.Context, key string, v *model.Forecast, ttl time.Duration)

	GetTide(ctx context.Context, key string) (*model.TidePrediction, bool)
	SetTide(ctx context.Context, key string, v *model.TidePrediction, ttl time.Duration)

	GetRadar(ctx context.Context) (*model.RainRadar, bool)
	SetRadar(ctx context.Context, v *model.RainRadar, ttl time.Duration)

	GetSnapshot(ctx context.Context, key string) (*model.WeatherSnapshot, bool)
	SetSnapshot(ctx context.Context, key string, v *model.WeatherSnapshot, ttl time.Duration)
}

type MarketProvider interface {
	GetMarkets(ctx context.Context, vsCurrency string, perPage int) ([]model.Coin, error)
}

type MarketCache interface {
	GetMarketSnapshot(ctx context.Context, key string) (*model.MarketSnapshot, bool)
	SetMarketSnapshot(ctx context.Context, key string, v *model.MarketSnapshot, ttl time.Duration)
}

type MarketBroadcaster interface {
	Broadcast(v *model.MarketSnapshot)
}
