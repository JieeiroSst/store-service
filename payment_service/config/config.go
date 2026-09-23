package config

import (
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	PayPal   PayPalConfig
	Payoneer PayoneerConfig
	Stripe   StripeConfig
	Wise     WiseConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
}

type PayPalConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Timeout      string
	WebhookID    string
}

type PayoneerConfig struct {
	BaseURL       string
	APIKey        string
	ProgramID     string
	Timeout       string
	WebhookSecret string
}

type StripeConfig struct {
	BaseURL       string
	SecretKey     string
	Timeout       string
	WebhookSecret string
}

type WiseConfig struct {
	BaseURL          string
	APIToken         string
	ProfileID        string
	Timeout          string
	WebhookPublicKey string
}

func (c PayPalConfig) TimeoutDuration() time.Duration   { return parseTimeout(c.Timeout) }
func (c PayoneerConfig) TimeoutDuration() time.Duration { return parseTimeout(c.Timeout) }
func (c StripeConfig) TimeoutDuration() time.Duration   { return parseTimeout(c.Timeout) }
func (c WiseConfig) TimeoutDuration() time.Duration     { return parseTimeout(c.Timeout) }

func parseTimeout(raw string) time.Duration {
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	return 10 * time.Second
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8090"),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "payment_service"),
		},
		PayPal: PayPalConfig{
			BaseURL:      getEnv("PAYPAL_BASE_URL", "https://api-m.sandbox.paypal.com"),
			ClientID:     getEnv("PAYPAL_CLIENT_ID", ""),
			ClientSecret: getEnv("PAYPAL_CLIENT_SECRET", ""),
			Timeout:      getEnv("PAYPAL_TIMEOUT", "10s"),
			WebhookID:    getEnv("PAYPAL_WEBHOOK_ID", ""),
		},
		Payoneer: PayoneerConfig{
			BaseURL:       getEnv("PAYONEER_BASE_URL", "https://api.sandbox.payoneer.com"),
			APIKey:        getEnv("PAYONEER_API_KEY", ""),
			ProgramID:     getEnv("PAYONEER_PROGRAM_ID", ""),
			Timeout:       getEnv("PAYONEER_TIMEOUT", "10s"),
			WebhookSecret: getEnv("PAYONEER_WEBHOOK_SECRET", ""),
		},
		Stripe: StripeConfig{
			BaseURL:       getEnv("STRIPE_BASE_URL", "https://api.stripe.com"),
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			Timeout:       getEnv("STRIPE_TIMEOUT", "10s"),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},
		Wise: WiseConfig{
			BaseURL:          getEnv("WISE_BASE_URL", "https://api.sandbox.transferwise.tech"),
			APIToken:         getEnv("WISE_API_TOKEN", ""),
			ProfileID:        getEnv("WISE_PROFILE_ID", ""),
			Timeout:          getEnv("WISE_TIMEOUT", "10s"),
			WebhookPublicKey: getEnv("WISE_WEBHOOK_PUBLIC_KEY", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
