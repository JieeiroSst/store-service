package config

import (
	"os"
	"time"
)

// Config holds doordash-service's own settings plus the base URLs of the
// sibling services it calls over HTTP to enrich orders with data it does
// not own (customer/driver identity, restaurant/menu catalog, payment
// authorization, notifications). See internal/domain/port.UserClient,
// RestaurantClient, PaymentClient and NotifierClient.
type Config struct {
	Server              ServerConfig
	Postgres            PostgresConfig
	UserService         ExternalServiceConfig
	RestaurantService   ExternalServiceConfig
	PaymentService      ExternalServiceConfig
	NotificationService ExternalServiceConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
}

type ExternalServiceConfig struct {
	BaseURL string
	Timeout string
}

func (e ExternalServiceConfig) TimeoutDuration() time.Duration {
	if d, err := time.ParseDuration(e.Timeout); err == nil {
		return d
	}
	return 5 * time.Second
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8086"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "doordash"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSLMODE", "disable") == "require",
		},
		UserService: ExternalServiceConfig{
			BaseURL: getEnv("USER_SERVICE_BASE_URL", "http://user-service-svc"),
			Timeout: getEnv("USER_SERVICE_TIMEOUT", "5s"),
		},
		RestaurantService: ExternalServiceConfig{
			BaseURL: getEnv("RESTAURANT_SERVICE_BASE_URL", "http://restaurant-service-svc"),
			Timeout: getEnv("RESTAURANT_SERVICE_TIMEOUT", "5s"),
		},
		PaymentService: ExternalServiceConfig{
			BaseURL: getEnv("PAYMENT_SERVICE_BASE_URL", "http://payment-service-svc"),
			Timeout: getEnv("PAYMENT_SERVICE_TIMEOUT", "5s"),
		},
		NotificationService: ExternalServiceConfig{
			BaseURL: getEnv("NOTIFICATION_SERVICE_BASE_URL", "http://notification-service-svc"),
			Timeout: getEnv("NOTIFICATION_SERVICE_TIMEOUT", "5s"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
