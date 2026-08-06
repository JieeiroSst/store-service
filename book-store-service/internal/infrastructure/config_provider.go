package infrastructure

import (
	"log"
	"os"

	"github.com/JIeeiroSst/bookStore-service/config"
	"github.com/JIeeiroSst/bookStore-service/pkg/consul"
	"github.com/joho/godotenv"
)

func newConfig() (*config.Config, error) {
	_ = godotenv.Load(".env")

	host := os.Getenv("HostConsul")
	key := os.Getenv("KeyConsul")
	service := os.Getenv("ServiceConsul")

	if host == "" || key == "" || service == "" {
		return config.FromEnv(), nil
	}

	cfg, err := consul.NewConfigConsul(host, key, service).ConnectConfigConsul()
	if err != nil || cfg == nil {
		log.Printf("consul config unavailable, falling back to env: %v", err)
		return config.FromEnv(), nil
	}
	return cfg, nil
}
