package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *queryResolver) CalcForecast(ctx context.Context, lat float64, lon float64) (*model.CalcForecast, error) {
	return restCall[*model.CalcForecast](ctx, r.Proxy, "calculate-service", rest.Request{Method: "GET", Path: "/api/v1/weather/forecast", Query: restQuery("lat", lat, "lon", lon)}, "")
}

func (r *queryResolver) CalcTide(ctx context.Context, station string) (*model.CalcTidePrediction, error) {
	return restCall[*model.CalcTidePrediction](ctx, r.Proxy, "calculate-service", rest.Request{Method: "GET", Path: "/api/v1/weather/tide", Query: restQuery("station", station)}, "")
}

func (r *queryResolver) CalcRadar(ctx context.Context) (*model.CalcRainRadar, error) {
	return restCall[*model.CalcRainRadar](ctx, r.Proxy, "calculate-service", rest.Request{Method: "GET", Path: "/api/v1/weather/radar"}, "")
}

func (r *queryResolver) CalcCurrentWeather(ctx context.Context, location string) (*model.CalcWeatherSnapshot, error) {
	return restCall[*model.CalcWeatherSnapshot](ctx, r.Proxy, "calculate-service", rest.Request{Method: "GET", Path: "/api/v1/weather/current", Query: restQuery("location", location)}, "")
}

func (r *queryResolver) CalcLocations(ctx context.Context) ([]*model.CalcTrackedLocation, error) {
	return restCall[[]*model.CalcTrackedLocation](ctx, r.Proxy, "calculate-service", rest.Request{Method: "GET", Path: "/api/v1/weather/locations"}, "")
}

func (r *queryResolver) CalcMarketSnapshot(ctx context.Context) (*model.CalcMarketSnapshot, error) {
	return restCall[*model.CalcMarketSnapshot](ctx, r.Proxy, "calculate-service", rest.Request{Method: "GET", Path: "/api/v1/market/snapshot"}, "")
}
