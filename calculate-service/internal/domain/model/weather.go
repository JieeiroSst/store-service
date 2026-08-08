package model

import "time"

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type ForecastPoint struct {
	Time               time.Time `json:"time"`
	TemperatureC       float64   `json:"temperature_c"`
	PrecipitationMm    float64   `json:"precipitation_mm"`
	WindSpeedKmh       float64   `json:"wind_speed_kmh"`
	WindDirectionDeg   float64   `json:"wind_direction_deg"`
	SurfacePressureHpa float64   `json:"surface_pressure_hpa"`
	PressureMslHpa     float64   `json:"pressure_msl_hpa"`
}

type Forecast struct {
	Coordinate  Coordinate      `json:"coordinate"`
	GeneratedAt time.Time       `json:"generated_at"`
	Source      string          `json:"source"`
	Hourly      []ForecastPoint `json:"hourly"`
}

type TideEventType string

const (
	TideHigh TideEventType = "high"
	TideLow  TideEventType = "low"
)

type TideEvent struct {
	Time    time.Time     `json:"time"`
	Type    TideEventType `json:"type"`
	HeightM float64       `json:"height_m"`
}

type TidePrediction struct {
	StationID string      `json:"station_id"`
	Source    string      `json:"source"`
	Events    []TideEvent `json:"events"`
}

type RadarFrame struct {
	Time    time.Time `json:"time"`
	TileURL string    `json:"tile_url"`
}

type RainRadar struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Host        string       `json:"host"`
	Past        []RadarFrame `json:"past"`
	Nowcast     []RadarFrame `json:"nowcast"`
	Source      string       `json:"source"`
}

type WeatherSnapshot struct {
	Location    string         `json:"location"`
	Coordinate  Coordinate     `json:"coordinate"`
	GeneratedAt time.Time      `json:"generated_at"`
	Current     *ForecastPoint `json:"current,omitempty"`
	NextTide    *TideEvent     `json:"next_tide,omitempty"`
	Radar       *RadarFrame    `json:"latest_radar,omitempty"`
}

type TrackedLocation struct {
	Name          string     `json:"name"`
	Coordinate    Coordinate `json:"coordinate"`
	TideStationID string     `json:"tide_station_id,omitempty"`
}
