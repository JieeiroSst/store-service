package config

import "os"

type Config struct {
	Server  ServerConfig
	Cache   CacheConfig
	Weather WeatherConfig
}

type ServerConfig struct {
	PortServer string
}

type CacheConfig struct {
	DNS string
}

type WeatherConfig struct {
	OpenMeteoBaseURL   string
	NoaaBaseURL        string
	RainViewerBaseURL  string
	RequestTimeoutSec  int
	RefreshIntervalMin int
	Locations          []LocationConfig
}

type LocationConfig struct {
	Name          string
	Lat           float64
	Lon           float64
	TideStationID string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortServer: getEnv("PORT_SERVER", "1239"),
		},
		Cache: CacheConfig{
			DNS: getEnv("CACHE_DNS", "localhost:6379"),
		},
		Weather: WeatherConfig{
			OpenMeteoBaseURL:   getEnv("OPEN_METEO_BASE_URL", "https://api.open-meteo.com/v1/forecast"),
			NoaaBaseURL:        getEnv("NOAA_BASE_URL", "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter"),
			RainViewerBaseURL:  getEnv("RAINVIEWER_BASE_URL", "https://api.rainviewer.com/public/weather-maps.json"),
			RequestTimeoutSec:  10,
			RefreshIntervalMin: 15,
			Locations: []LocationConfig{
				{Name: "ho-chi-minh-city", Lat: 10.7769, Lon: 106.7009},
				{Name: "vung-tau", Lat: 10.3460, Lon: 107.0843},
				{Name: "da-nang", Lat: 16.0544, Lon: 108.2022},
				{Name: "new-york-the-battery", Lat: 40.7003, Lon: -74.0142, TideStationID: "8518750"},
			},
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
