package config

import (
	"encoding/json"
	"fmt"
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

// InitializeConfiguration loads configuration exclusively from Consul KV.
// HostConsul/KeyConsul/ServiceConsul bootstrap coordinates come from the
// .env file; there is no fallback to reading individual config values from
// the environment.
func InitializeConfiguration(ecosystem string) (*Config, error) {
	_ = godotenv.Load(ecosystem)

	host := os.Getenv("HostConsul")
	key := os.Getenv("KeyConsul")
	service := os.Getenv("ServiceConsul")

	data, err := consul.NewConfigConsul(host, key, service).ConnectConfigConsul()
	if err != nil {
		return nil, fmt.Errorf("consul config unavailable: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("consul returned no config for key %q", key)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse consul config: %w", err)
	}
	return &cfg, nil
}
