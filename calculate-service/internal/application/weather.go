package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/JIeeiroSst/calculate-service/common"
	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
)

const (
	forecastTTL = 15 * time.Minute
	tideTTL     = 6 * time.Hour
	radarTTL    = 5 * time.Minute
	snapshotTTL = 20 * time.Minute
)

const (
	vnMinLat, vnMaxLat = -10.0, 24.0
	vnMinLon, vnMaxLon = 90.0, 150.0
)

type weatherService struct {
	forecast  port.ForecastProvider
	tide      port.TideProvider
	radar     port.RadarProvider
	cache     port.WeatherCache
	locations []model.TrackedLocation
}

func NewWeatherService(
	forecast port.ForecastProvider,
	tide port.TideProvider,
	radar port.RadarProvider,
	cache port.WeatherCache,
	locations []model.TrackedLocation,
) port.WeatherService {
	return &weatherService{
		forecast:  forecast,
		tide:      tide,
		radar:     radar,
		cache:     cache,
		locations: locations,
	}
}

func forecastCacheKey(c model.Coordinate) string {
	return fmt.Sprintf("weather:forecast:%.3f:%.3f", c.Lat, c.Lon)
}

func tideCacheKey(stationID string) string {
	return "weather:tide:" + stationID
}

func snapshotCacheKey(location string) string {
	return "weather:snapshot:" + location
}

func (s *weatherService) GetForecast(ctx context.Context, coord model.Coordinate) (*model.Forecast, error) {
	if coord.Lat < -90 || coord.Lat > 90 || coord.Lon < -180 || coord.Lon > 180 {
		return nil, common.ErrInvalidCoordinate
	}

	key := forecastCacheKey(coord)
	if cached, ok := s.cache.GetForecast(ctx, key); ok {
		return cached, nil
	}

	f, err := s.forecast.GetForecast(ctx, coord)
	if err != nil {
		logger.Error(ctx, "GetForecast upstream error %v", err)
		return nil, fmt.Errorf("%w: %v", common.ErrUpstreamUnavailable, err)
	}
	s.cache.SetForecast(ctx, key, f, forecastTTL)
	return f, nil
}

func (s *weatherService) GetTide(ctx context.Context, stationID string) (*model.TidePrediction, error) {
	if stationID == "" {
		return nil, fmt.Errorf("%w: station is required", common.ErrInvalidCoordinate)
	}

	key := tideCacheKey(stationID)
	if cached, ok := s.cache.GetTide(ctx, key); ok {
		return cached, nil
	}

	t, err := s.tide.GetTidePredictions(ctx, stationID)
	if err != nil {
		logger.Error(ctx, "GetTide upstream error %v", err)
		return nil, fmt.Errorf("%w: %v", common.ErrUpstreamUnavailable, err)
	}
	s.cache.SetTide(ctx, key, t, tideTTL)
	return t, nil
}

func (s *weatherService) GetRadar(ctx context.Context) (*model.RainRadar, error) {
	if cached, ok := s.cache.GetRadar(ctx); ok {
		return cached, nil
	}

	r, err := s.radar.GetLatestRadar(ctx)
	if err != nil {
		logger.Error(ctx, "GetRadar upstream error %v", err)
		return nil, fmt.Errorf("%w: %v", common.ErrUpstreamUnavailable, err)
	}
	s.cache.SetRadar(ctx, r, radarTTL)
	return r, nil
}

func (s *weatherService) GetCurrentConditions(ctx context.Context, location string) (*model.WeatherSnapshot, error) {
	if cached, ok := s.cache.GetSnapshot(ctx, snapshotCacheKey(location)); ok {
		return cached, nil
	}
	return nil, common.ErrLocationNotTracked
}

func (s *weatherService) ListTrackedLocations(ctx context.Context) []model.TrackedLocation {
	return s.locations
}

func (s *weatherService) RefreshTrackedLocations(ctx context.Context) error {
	radar, radarErr := s.radar.GetLatestRadar(ctx)
	if radarErr != nil {
		logger.Error(ctx, "RefreshTrackedLocations radar error %v", radarErr)
	} else {
		s.cache.SetRadar(ctx, radar, radarTTL)
	}

	var lastErr error
	for _, loc := range s.locations {
		forecast, err := s.forecast.GetForecast(ctx, loc.Coordinate)
		if err != nil {
			logger.Error(ctx, "RefreshTrackedLocations forecast for %s error %v", loc.Name, err)
			lastErr = err
			continue
		}
		s.cache.SetForecast(ctx, forecastCacheKey(loc.Coordinate), forecast, forecastTTL)

		snapshot := &model.WeatherSnapshot{
			Location:    loc.Name,
			Coordinate:  loc.Coordinate,
			GeneratedAt: time.Now().UTC(),
			Current:     nearestForecastPoint(forecast.Hourly),
		}

		if loc.TideStationID != "" {
			tide, err := s.tide.GetTidePredictions(ctx, loc.TideStationID)
			if err != nil {
				logger.Error(ctx, "RefreshTrackedLocations tide for %s error %v", loc.Name, err)
			} else {
				s.cache.SetTide(ctx, tideCacheKey(loc.TideStationID), tide, tideTTL)
				snapshot.NextTide = nextTideEvent(tide.Events)
			}
		}

		if radarErr == nil {
			snapshot.Radar = latestRadarFrame(radar)
		}

		s.cache.SetSnapshot(ctx, snapshotCacheKey(loc.Name), snapshot, snapshotTTL)
	}
	return lastErr
}

func nearestForecastPoint(points []model.ForecastPoint) *model.ForecastPoint {
	if len(points) == 0 {
		return nil
	}
	now := time.Now().UTC()
	best := &points[0]
	bestDiff := absDuration(points[0].Time.Sub(now))
	for i := 1; i < len(points); i++ {
		if diff := absDuration(points[i].Time.Sub(now)); diff < bestDiff {
			bestDiff = diff
			best = &points[i]
		}
	}
	return best
}

func nextTideEvent(events []model.TideEvent) *model.TideEvent {
	if len(events) == 0 {
		return nil
	}
	sorted := make([]model.TideEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	now := time.Now().UTC()
	for i := range sorted {
		if sorted[i].Time.After(now) {
			return &sorted[i]
		}
	}
	return &sorted[len(sorted)-1]
}

func latestRadarFrame(r *model.RainRadar) *model.RadarFrame {
	if r == nil || len(r.Past) == 0 {
		return nil
	}
	f := r.Past[len(r.Past)-1]
	return &f
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
