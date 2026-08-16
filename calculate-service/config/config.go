package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig  `json:"server"`
	Cache   CacheConfig   `json:"cache"`
	Weather WeatherConfig `json:"weather"`
	Market  MarketConfig  `json:"market"`
}

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

type CacheConfig struct {
	DNS string `json:"dns"`
}

type WeatherConfig struct {
	OpenMeteoBaseURL   string           `json:"open_meteo_base_url"`
	NoaaBaseURL        string           `json:"noaa_base_url"`
	RainViewerBaseURL  string           `json:"rain_viewer_base_url"`
	RequestTimeoutSec  int              `json:"request_timeout_sec"`
	RefreshIntervalMin int              `json:"refresh_interval_min"`
	Locations          []LocationConfig `json:"locations"`
}

type LocationConfig struct {
	Name          string  `json:"name"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	TideStationID string  `json:"tide_station_id"`
}

type MarketConfig struct {
	CoinGeckoBaseURL   string `json:"coingecko_base_url"`
	VsCurrency         string `json:"vs_currency"`
	PerPage            int    `json:"per_page"`
	RequestTimeoutSec  int    `json:"request_timeout_sec"`
	RefreshIntervalSec int    `json:"refresh_interval_sec"`
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func ReadFileEnv(dir string) (*Dir, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	data := &Dir{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}
	return data, nil
}
