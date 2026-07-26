package config

import (
	"encoding/json"
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
	HTTPPort string `json:"httpPort"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresqlHost"`
	PostgresqlPort     string `json:"postgresqlPort"`
	PostgresqlUser     string `json:"postgresqlUser"`
	PostgresqlPassword string `json:"postgresqlPassword"`
	PostgresqlDbname   string `json:"postgresqlDbname"`
	PostgresqlSSLMode  bool   `json:"postgresqlSSLMode"`
}

type BillingConfig struct {
	RenewalCheckInterval time.Duration `json:"renewalCheckInterval"`
}

// GatewayConfig points at the outbound payment gateway. Charges are
// delegated to integrated-payment-service over HTTP rather than integrating
// a provider (Stripe/VNPay/...) directly in this service.
type GatewayConfig struct {
	IntegratedPaymentServiceURL string `json:"integratedPaymentServiceUrl"`
}

func InitializeConfiguration(ecosystem string) (*Config, error) {
	_ = godotenv.Load(ecosystem)

	host := os.Getenv("HostConsul")
	key := os.Getenv("KeyConsul")
	service := os.Getenv("ServiceConsul")

	consulClient := consul.NewConfigConsul(host, key, service)
	data, err := consulClient.ConnectConfigConsul()
	if err != nil || data == nil {
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
			HTTPPort: getEnv("HTTP_PORT", "8080"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("PG_HOST", "localhost"),
			PostgresqlPort:     getEnv("PG_PORT", "5432"),
			PostgresqlUser:     getEnv("PG_USER", "postgres"),
			PostgresqlPassword: getEnv("PG_PASSWORD", ""),
			PostgresqlDbname:   getEnv("PG_DB", "automatic_payment"),
		},
		Billing: BillingConfig{
			RenewalCheckInterval: getEnvDuration("RENEWAL_CHECK_INTERVAL", 24*time.Hour),
		},
		Gateway: GatewayConfig{
			IntegratedPaymentServiceURL: getEnv("INTEGRATED_PAYMENT_SERVICE_URL", "http://localhost:8080"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
