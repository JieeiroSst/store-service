package config

import "os"

type Config struct {
	Env      string
	Temporal TemporalConfig
	Services ServicesConfig
	OrderDB  string
}

type TemporalConfig struct {
	HostPort  string
	Namespace string
}

type ServicesConfig struct {
	PaymentBaseURL      string
	PaymentAPIKey       string
	InventoryBaseURL    string
	InventoryAPIKey     string
	ShippingBaseURL     string
	ShippingAPIKey      string
	NotificationBaseURL string
	NotificationAPIKey  string
}

func Load() *Config {
	return &Config{
		Env:     getEnv("APP_ENV", "development"),
		OrderDB: getEnv("ORDER_DB_DSN", ""),
		Temporal: TemporalConfig{
			HostPort:  getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
			Namespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		},
		Services: ServicesConfig{
			PaymentBaseURL:      getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081"),
			PaymentAPIKey:       getEnv("PAYMENT_SERVICE_API_KEY", ""),
			InventoryBaseURL:    getEnv("INVENTORY_SERVICE_URL", "http://localhost:8082"),
			InventoryAPIKey:     getEnv("INVENTORY_SERVICE_API_KEY", ""),
			ShippingBaseURL:     getEnv("SHIPPING_SERVICE_URL", "http://localhost:8083"),
			ShippingAPIKey:      getEnv("SHIPPING_SERVICE_API_KEY", ""),
			NotificationBaseURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8084"),
			NotificationAPIKey:  getEnv("NOTIFICATION_SERVICE_API_KEY", getEnv("INTERNAL_API_TOKEN", "")),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
