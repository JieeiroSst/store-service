package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	Billing  BillingConfig  `json:"billing"`
	Gateway  GatewayConfig  `json:"gateway"`
}

type ServerConfig struct {
	HTTPPort string `json:"http_port"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresql_host"`
	PostgresqlPort     string `json:"postgresql_port"`
	PostgresqlUser     string `json:"postgresql_user"`
	PostgresqlPassword string `json:"postgresql_password"`
	PostgresqlDbname   string `json:"postgresql_dbname"`
	PostgresqlSSLMode  bool   `json:"postgresql_ssl_mode"`
}

type BillingConfig struct {
	RenewalCheckInterval time.Duration `json:"renewal_check_interval"`
}

// GatewayConfig points at the outbound payment gateway. Charges are
// delegated to integrated-payment-service over HTTP rather than integrating
// a provider (Stripe/VNPay/...) directly in this service.
type GatewayConfig struct {
	IntegratedPaymentServiceURL string `json:"integrated_payment_service_url"`
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
