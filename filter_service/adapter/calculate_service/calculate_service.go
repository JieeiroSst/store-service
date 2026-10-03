package calculate_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "calculate-service"

const DefaultBaseURL = "http://calculate-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Forecast(ctx context.Context, lat float64, lon float64) (*model.CalcForecast, error) {
	path := "/api/v1/weather/forecast"
	q := url.Values{}
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	h := http.Header{}
	var out *model.CalcForecast
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Tide(ctx context.Context, station *string) (*model.CalcTidePrediction, error) {
	path := "/api/v1/weather/tide"
	q := url.Values{}
	if station != nil {
		q.Set("station", *station)
	}
	h := http.Header{}
	var out *model.CalcTidePrediction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Radar(ctx context.Context) (*model.CalcRainRadar, error) {
	path := "/api/v1/weather/radar"
	q := url.Values{}
	h := http.Header{}
	var out *model.CalcRainRadar
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Current(ctx context.Context, location *string) (*model.CalcWeatherSnapshot, error) {
	path := "/api/v1/weather/current"
	q := url.Values{}
	if location != nil {
		q.Set("location", *location)
	}
	h := http.Header{}
	var out *model.CalcWeatherSnapshot
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Locations(ctx context.Context) ([]*model.CalcTrackedLocation, error) {
	path := "/api/v1/weather/locations"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CalcTrackedLocation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MarketSnapshot(ctx context.Context) (*model.CalcMarketSnapshot, error) {
	path := "/api/v1/market/snapshot"
	q := url.Values{}
	h := http.Header{}
	var out *model.CalcMarketSnapshot
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
