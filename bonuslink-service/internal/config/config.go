package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(Load),
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	RabbitMQ RabbitMQConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Env     string
	Port    int
	Name    string
	Version string
}

type PostgresConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RabbitMQConfig struct {
	Enabled    bool
	Host       string
	Port       string
	User       string
	Password   string
	VHost      string
	Exchange   string
	Queue      string
	RoutingKey string
	Prefetch   int
}

type LoggerConfig struct {
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

// URL builds the AMQP connection string.
func (c RabbitMQConfig) URL() string {
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(c.User, c.Password),
		Host:   c.Host + ":" + c.Port,
		Path:   "/" + c.VHost,
	}
	return u.String()
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("config: APP_PORT must be integer: %w", err)
	}

	return &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Port:    port,
			Name:    getEnv("APP_NAME", "bonuslink-service"),
			Version: getEnv("APP_VERSION", "0.0.0"),
		},
		Postgres: PostgresConfig{
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            getEnv("POSTGRES_PORT", "5432"),
			User:            getEnv("COM_POSTGRES_USERNAME", "postgres"),
			Password:        getEnv("COM_POSTGRES_PASSWORD", "postgres"),
			Database:        getEnv("POSTGRES_DATABASE", "bonuslink_service"),
			SSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		RabbitMQ: RabbitMQConfig{
			Enabled:    getEnv("RABBITMQ_ENABLED", "true") == "true",
			Host:       getEnv("RABBITMQ_HOST", "localhost"),
			Port:       getEnv("RABBITMQ_PORT", "5672"),
			User:       getEnv("COM_RABBITMQ_USERNAME", "guest"),
			Password:   getEnv("COM_RABBITMQ_PASSWORD", "guest"),
			VHost:      getEnv("RABBITMQ_VHOST", ""),
			Exchange:   getEnv("RABBITMQ_EXCHANGE", "referral.events"),
			Queue:      getEnv("RABBITMQ_QUEUE", "bonuslink.reward"),
			RoutingKey: getEnv("RABBITMQ_ROUTING_KEY", "referral.reward.granted"),
			Prefetch:   getEnvAsInt("RABBITMQ_PREFETCH", 10),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			FilePath:   getEnv("LOG_FILE_PATH", ""),
			MaxSizeMB:  getEnvAsInt("LOG_MAX_SIZE_MB", 100),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 7),
			MaxAgeDays: getEnvAsInt("LOG_MAX_AGE_DAYS", 30),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
