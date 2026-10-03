package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CalcQuery() generated.CalcQueryResolver { return &calcQueryResolver{r} }

type calcQueryResolver struct{ *Resolver }

func (r *calcQueryResolver) Forecast(ctx context.Context, obj *model.CalcQuery, lat float64, lon float64) (*model.CalcForecast, error) {
	return r.Clients.CalculateService.Forecast(ctx, lat, lon)
}

func (r *calcQueryResolver) Tide(ctx context.Context, obj *model.CalcQuery, station *string) (*model.CalcTidePrediction, error) {
	return r.Clients.CalculateService.Tide(ctx, station)
}

func (r *calcQueryResolver) Radar(ctx context.Context, obj *model.CalcQuery) (*model.CalcRainRadar, error) {
	return r.Clients.CalculateService.Radar(ctx)
}

func (r *calcQueryResolver) Current(ctx context.Context, obj *model.CalcQuery, location *string) (*model.CalcWeatherSnapshot, error) {
	return r.Clients.CalculateService.Current(ctx, location)
}

func (r *calcQueryResolver) Locations(ctx context.Context, obj *model.CalcQuery) ([]*model.CalcTrackedLocation, error) {
	return r.Clients.CalculateService.Locations(ctx)
}

func (r *calcQueryResolver) MarketSnapshot(ctx context.Context, obj *model.CalcQuery) (*model.CalcMarketSnapshot, error) {
	return r.Clients.CalculateService.MarketSnapshot(ctx)
}
