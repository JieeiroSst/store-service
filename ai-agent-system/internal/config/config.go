package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	OllamaBaseURL string
	OllamaModel   string
	MySQLDSN      string
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		OllamaBaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:   os.Getenv("OLLAMA_MODEL"),
		MySQLDSN:      os.Getenv("MYSQL_DSN"),
	}

	if cfg.OllamaModel == "" {
		return Config{}, fmt.Errorf("config: OLLAMA_MODEL is required")
	}
	if cfg.MySQLDSN == "" {
		return Config{}, fmt.Errorf("config: MYSQL_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
