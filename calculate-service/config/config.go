package config

import (
	"os"

	"github.com/joho/godotenv"
)

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
