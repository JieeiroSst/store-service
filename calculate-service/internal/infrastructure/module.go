package infrastructure

import (
	"net/http"

	"github.com/JIeeiroSst/calculate-service/config"
	httpadapter "github.com/JIeeiroSst/calculate-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/primary/ws"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/cache"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/coingecko"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/noaa"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/openmeteo"
	"github.com/JIeeiroSst/calculate-service/internal/adapter/secondary/rainviewer"
	"github.com/JIeeiroSst/calculate-service/internal/application"
	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
	"github.com/JIeeiroSst/calculate-service/internal/infrastructure/server"
	"github.com/JIeeiroSst/calculate-service/internal/infrastructure/worker"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),
	fx.Provide(NewHTTPClient),
	fx.Provide(newRedisClient),

	fx.Provide(newForecastProvider), // port.ForecastProvider (Open-Meteo)
	fx.Provide(newTideProvider),     // port.TideProvider (NOAA)
	fx.Provide(newRadarProvider),    // port.RadarProvider (RainViewer)
	fx.Provide(newWeatherCache),     // port.WeatherCache (Redis)
	fx.Provide(newTrackedLocations), // []model.TrackedLocation

	fx.Provide(newMarketProvider),      // port.MarketProvider (CoinGecko)
	fx.Provide(newMarketCache),         // port.MarketCache (Redis)
	fx.Provide(ws.NewHub),              // *ws.Hub
	fx.Provide(newMarketBroadcaster),   // port.MarketBroadcaster
	fx.Provide(newMarketServiceConfig), // application.MarketServiceConfig

	application.Module, // port.WeatherService, port.MarketService

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
	fx.Invoke(worker.New),
	fx.Invoke(worker.NewMarket),
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

func newRedisClient(cfg *config.Config) *redis.Client {
	return cache.NewRedisClient(cfg.Cache.DNS)
}

func newWeatherCache(rdb *redis.Client) port.WeatherCache {
	return cache.NewRedisCache(rdb)
}

func newMarketProvider(cfg *config.Config, httpClient *http.Client) port.MarketProvider {
	return coingecko.NewClient(cfg.Market.CoinGeckoBaseURL, httpClient)
}

func newMarketCache(rdb *redis.Client) port.MarketCache {
	return cache.NewRedisMarketCache(rdb)
}

func newMarketBroadcaster(hub *ws.Hub) port.MarketBroadcaster {
	return hub
}

func newMarketServiceConfig(cfg *config.Config) application.MarketServiceConfig {
	return application.MarketServiceConfig{
		VsCurrency: cfg.Market.VsCurrency,
		PerPage:    cfg.Market.PerPage,
	}
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
