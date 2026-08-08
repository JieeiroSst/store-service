package infrastructure

import (
	"context"
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/calculate-service/config"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/logger"
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
		logger.Error(context.Background(), "consul config unavailable, falling back to env: %v", err)
		return config.FromEnv(), nil
	}

	var cfg config.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		logger.Error(context.Background(), "consul config unmarshal error, falling back to env: %v", err)
		return config.FromEnv(), nil
	}
	return &cfg, nil
}
