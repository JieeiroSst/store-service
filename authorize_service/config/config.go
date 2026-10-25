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
	Server    ServerConfig    `json:"server"`
	Postgres  PostgresConfig  `json:"postgres"`
	Cache     CacheConfig     `json:"cache"`
	Secret    SecretConfig    `json:"secret"`
	Bootstrap BootstrapConfig `json:"bootstrap"`
	MFA       MFAConfig       `json:"mfa"`
}

type MFAConfig struct {
	Issuer        string `json:"issuer"`
	EncryptionKey string `json:"encryption_key"`
}

func (m MFAConfig) IssuerName() string {
	if m.Issuer == "" {
		return "Keiki"
	}
	return m.Issuer
}

type BootstrapConfig struct {
	SuperAdminUserIDs []string `json:"super_admin_user_ids"`
}

type ServerConfig struct {
	PortHttpServer string `json:"port_http_server"`
	PortGrpcServer string `json:"port_grpc_server"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresql_host"`
	PostgresqlPort     string `json:"postgresql_port"`
	PostgresqlUser     string `json:"postgresql_user"`
	PostgresqlPassword string `json:"postgresql_password"`
	PostgresqlDbname   string `json:"postgresql_dbname"`
	PostgresqlSSLMode  bool   `json:"postgresql_ssl_mode"`
}

type CacheConfig struct {
	Host string `json:"host"`
}

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
}

// InitializeConfiguration loads configuration from Consul KV.
// HostConsul/KeyConsul/ServiceConsul bootstrap coordinates come from the
// .env file; the only value read from the environment is MFA_ENCRYPTION_KEY,
// so the TOTP encryption key can live in the secret-managed .env instead of Consul.
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
	if v, ok := os.LookupEnv("MFA_ENCRYPTION_KEY"); ok && v != "" {
		cfg.MFA.EncryptionKey = v
	}
	return &cfg, nil
}
