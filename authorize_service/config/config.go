package config

import (
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

// Config holds all service configuration.
type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	Cache    CacheConfig    `json:"cache"`
	Secret   SecretConfig   `json:"secret"`
}

type ServerConfig struct {
	PortHttpServer string `json:"portHttpServer"`
	PortGrpcServer string `json:"portGrpcServer"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresqlHost"`
	PostgresqlPort     string `json:"postgresqlPort"`
	PostgresqlUser     string `json:"postgresqlUser"`
	PostgresqlPassword string `json:"postgresqlPassword"`
	PostgresqlDbname   string `json:"postgresqlDbname"`
	PostgresqlSSLMode  bool   `json:"postgresqlSSLMode"`
}

type CacheConfig struct {
	Host string `json:"host"`
}

type SecretConfig struct {
	JwtSecretKey string `json:"jwtSecretKey"`
}

// InitializeConfiguration loads config from Consul KV; falls back to .env file.
func InitializeConfiguration(ecosystem string) (*Config, error) {
	_ = godotenv.Load(ecosystem)

	host := os.Getenv("HostConsul")
	key := os.Getenv("KeyConsul")
	service := os.Getenv("ServiceConsul")

	consulClient := consul.NewConfigConsul(host, key, service)
	data, err := consulClient.ConnectConfigConsul()
	if err != nil || data == nil {
		// Consul unavailable — build config from environment variables.
		return configFromEnv(), nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func configFromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("HTTP_PORT", "8080"),
			PortGrpcServer: getEnv("GRPC_PORT", "9090"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("PG_HOST", "localhost"),
			PostgresqlPort:     getEnv("PG_PORT", "5432"),
			PostgresqlUser:     getEnv("PG_USER", "postgres"),
			PostgresqlPassword: getEnv("PG_PASSWORD", ""),
			PostgresqlDbname:   getEnv("PG_DB", "authorize"),
		},
		Cache: CacheConfig{
			Host: getEnv("REDIS_HOST", "localhost:6379"),
		},
		Secret: SecretConfig{
			JwtSecretKey: getEnv("JWT_SECRET", "secret"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
