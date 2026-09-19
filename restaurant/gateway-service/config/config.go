package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig  `json:"server"`
	Gateway GatewayConfig `json:"gateway"`
}

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

// GatewayConfig holds the base URLs of the three downstream services the
// gateway fronts, matching the drawio's outgoing edges from gateway-service.
type GatewayConfig struct {
	ConsumerURL   string `json:"consumer_url"`
	AccountingURL string `json:"accounting_url"`
	DeliveryURL   string `json:"delivery_url"`
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
