package port

import (
	"context"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
)

type WeatherService interface {
	GetForecast(ctx context.Context, coord model.Coordinate) (*model.Forecast, error)
	GetTide(ctx context.Context, stationID string) (*model.TidePrediction, error)
	GetRadar(ctx context.Context) (*model.RainRadar, error)
	GetCurrentConditions(ctx context.Context, location string) (*model.WeatherSnapshot, error)
	ListTrackedLocations(ctx context.Context) []model.TrackedLocation
	RefreshTrackedLocations(ctx context.Context) error
}

type MarketService interface {
	GetSnapshot(ctx context.Context) (*model.MarketSnapshot, error)
	RefreshMarkets(ctx context.Context) error
}
