package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
)

const timeLayout = "2006-01-02T15:04"

const hourlyParams = "temperature_2m,precipitation,windspeed_10m,winddirection_10m,surface_pressure,pressure_msl"

type client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient builds a ForecastProvider calling the Open-Meteo forecast
// endpoint, e.g. baseURL="https://api.open-meteo.com/v1/forecast".
func NewClient(baseURL string, httpClient *http.Client) port.ForecastProvider {
	return &client{baseURL: baseURL, httpClient: httpClient}
}

type forecastResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hourly    struct {
		Time             []string  `json:"time"`
		Temperature2m    []float64 `json:"temperature_2m"`
		Precipitation    []float64 `json:"precipitation"`
		WindSpeed10m     []float64 `json:"windspeed_10m"`
		WindDirection10m []float64 `json:"winddirection_10m"`
		SurfacePressure  []float64 `json:"surface_pressure"`
		PressureMsl      []float64 `json:"pressure_msl"`
	} `json:"hourly"`
}

func (c *client) GetForecast(ctx context.Context, coord model.Coordinate) (*model.Forecast, error) {
	q := url.Values{}
	q.Set("latitude", strconv.FormatFloat(coord.Lat, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(coord.Lon, 'f', -1, 64))
	q.Set("hourly", hourlyParams)
	q.Set("timezone", "UTC")

	reqURL := c.baseURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo: unexpected status %d", resp.StatusCode)
	}

	var body forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	hourly := make([]model.ForecastPoint, 0, len(body.Hourly.Time))
	for i, ts := range body.Hourly.Time {
		t, err := time.Parse(timeLayout, ts)
		if err != nil {
			continue
		}
		point := model.ForecastPoint{Time: t}
		if i < len(body.Hourly.Temperature2m) {
			point.TemperatureC = body.Hourly.Temperature2m[i]
		}
		if i < len(body.Hourly.Precipitation) {
			point.PrecipitationMm = body.Hourly.Precipitation[i]
		}
		if i < len(body.Hourly.WindSpeed10m) {
			point.WindSpeedKmh = body.Hourly.WindSpeed10m[i]
		}
		if i < len(body.Hourly.WindDirection10m) {
			point.WindDirectionDeg = body.Hourly.WindDirection10m[i]
		}
		if i < len(body.Hourly.SurfacePressure) {
			point.SurfacePressureHpa = body.Hourly.SurfacePressure[i]
		}
		if i < len(body.Hourly.PressureMsl) {
			point.PressureMslHpa = body.Hourly.PressureMsl[i]
		}
		hourly = append(hourly, point)
	}

	return &model.Forecast{
		Coordinate:  coord,
		GeneratedAt: time.Now().UTC(),
		Source:      "open-meteo",
		Hourly:      hourly,
	}, nil
}
