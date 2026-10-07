package infrastructure

import (
	"encoding/json"
	"log"
	"os"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"

	"github.com/JIeeiroSst/ekyc-service/config"
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
	env := config.FromEnv()
	if _, ok := os.LookupEnv("INTERNAL_TOKENS"); ok {
		cfg.Auth.InternalTokens = env.Auth.InternalTokens
	}
	if _, ok := os.LookupEnv("USER_SERVICE_TOKEN"); ok {
		cfg.UserService.Token = env.UserService.Token
	}
	return &cfg, nil
}
