package infrastructure

import (
	"encoding/json"
	"log"
	"os"

	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/utils/consul"
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

	raw, err := consul.NewConfigConsul(host, key, service).ConnectConfigConsul()
	if err != nil || raw == nil {
		log.Printf("consul config unavailable, falling back to env: %v", err)
		return config.FromEnv(), nil
	}

	var cfg config.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		log.Printf("failed to parse consul config, falling back to env: %v", err)
		return config.FromEnv(), nil
	}
	return &cfg, nil
}
