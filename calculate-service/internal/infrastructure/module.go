package infrastructure

import (
	"net/http"

	"github.com/JIeeiroSst/calculate-service/config"
	httpadapter "github.com/JIeeiroSst/calculate-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/cache"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/noaa"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/openmeteo"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/rainviewer"
	"github.com/JIeeiroSst/calculate-service/internal/application"
	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
	"github.com/JIeeiroSst/calculate-service/internal/infrastructure/server"
	"github.com/JIeeiroSst/calculate-service/internal/infrastructure/worker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),
	fx.Provide(NewHTTPClient),

	fx.Provide(newForecastProvider), // port.ForecastProvider (Open-Meteo)
	fx.Provide(newTideProvider),     // port.TideProvider (NOAA)
	fx.Provide(newRadarProvider),    // port.RadarProvider (RainViewer)
	fx.Provide(newWeatherCache),     // port.WeatherCache (Redis)
	fx.Provide(newTrackedLocations), // []model.TrackedLocation

	application.Module, // port.WeatherService

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
	fx.Invoke(worker.New),
)

func newForecastProvider(cfg *config.Config, httpClient *http.Client) port.ForecastProvider {
	return openmeteo.NewClient(cfg.Weather.OpenMeteoBaseURL, httpClient)
}

func newTideProvider(cfg *config.Config, httpClient *http.Client) port.TideProvider {
	return noaa.NewClient(cfg.Weather.NoaaBaseURL, httpClient)
}

func newRadarProvider(cfg *config.Config, httpClient *http.Client) port.RadarProvider {
	return rainviewer.NewClient(cfg.Weather.RainViewerBaseURL, httpClient)
}

func newWeatherCache(cfg *config.Config) port.WeatherCache {
	return cache.NewRedisCache(cfg.Cache.DNS)
}

func newTrackedLocations(cfg *config.Config) []model.TrackedLocation {
	locations := make([]model.TrackedLocation, 0, len(cfg.Weather.Locations))
	for _, l := range cfg.Weather.Locations {
		locations = append(locations, model.TrackedLocation{
			Name:          l.Name,
			Coordinate:    model.Coordinate{Lat: l.Lat, Lon: l.Lon},
			TideStationID: l.TideStationID,
		})
	}
	return locations
}
