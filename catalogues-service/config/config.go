package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
}

type ServerConfig struct {
	HTTPPort        string
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		Server: ServerConfig{
			HTTPPort:        get("PORT", "8080"),
			ShutdownTimeout: durationOr(get("SHUTDOWN_TIMEOUT", "15s"), 15*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     get("POSTGRES_HOST", "localhost"),
			Port:     get("POSTGRES_PORT", "5432"),
			User:     get("POSTGRES_USER", "postgres"),
			Password: get("POSTGRES_PASSWORD", ""),
			DBName:   get("POSTGRES_DATABASE", "catalogues_service"),
			SSLMode:  get("POSTGRES_SSLMODE", "disable"),
		},
	}
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func durationOr(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return def
}
