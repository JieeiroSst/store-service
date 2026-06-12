package config

import (
	"fmt"
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
	MySQL    MySQLConfig
	DeepLink DeepLinkConfig
	Referral ReferralConfig
	Logger   LoggerConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Env     string
	Port    int
	Name    string
	Version string
}

type MySQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type DeepLinkConfig struct {
	PublicURL    string
	AppStoreURL  string
	PlayStoreURL string
}

type ReferralConfig struct {
	TTLDays   int
	MaxPerDay int
}

type LoggerConfig struct {
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

type RedisConfig struct {
	Endpoint      string
	Port          string
	Username      string
	Password      string
	Database      int
	PoolSize      int
	MinConnection int
	Timeout       time.Duration
	TLS           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("config: APP_PORT must be integer: %w", err)
	}

	ttl, err := strconv.Atoi(getEnv("REFERRAL_TTL_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("config: REFERRAL_TTL_DAYS must be integer: %w", err)
	}

	maxPerDay, err := strconv.Atoi(getEnv("MAX_REFERRAL_PER_DAY", "50"))
	if err != nil {
		return nil, fmt.Errorf("config: MAX_REFERRAL_PER_DAY must be integer: %w", err)
	}

	logMaxSize, err := strconv.Atoi(getEnv("LOG_MAX_SIZE_MB", "100"))
	if err != nil {
		return nil, fmt.Errorf("config: LOG_MAX_SIZE_MB must be integer: %w", err)
	}

	logMaxBackups, err := strconv.Atoi(getEnv("LOG_MAX_BACKUPS", "7"))
	if err != nil {
		return nil, fmt.Errorf("config: LOG_MAX_BACKUPS must be integer: %w", err)
	}

	logMaxAge, err := strconv.Atoi(getEnv("LOG_MAX_AGE_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("config: LOG_MAX_AGE_DAYS must be integer: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Port:    port,
			Name:    getEnv("APP_NAME", "referral-service"),
			Version: getEnv("APP_VERSION", "0.0.0"),
		},
		MySQL: MySQLConfig{
			Host:            getEnv("MYSQL_HOST", "localhost"),
			Port:            getEnv("MYSQL_PORT", "3306"),
			User:            getEnv("COM_MYSQL_USERNAME", "root"),
			Password:        getEnv("COM_MYSQL_PASSWORD", "root"),
			Database:        getEnv("MYSQL_DATABASE", "referral_service"),
			MaxOpenConns:    getEnvAsInt("MYSQL_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getEnvAsInt("MYSQL_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		DeepLink: DeepLinkConfig{
			PublicURL:    getEnv("APP_PUBLIC_URL", fmt.Sprintf("http://localhost:%d", port)),
			AppStoreURL:  getEnv("APP_STORE_URL", ""),
			PlayStoreURL: getEnv("PLAY_STORE_URL", ""),
		},
		Referral: ReferralConfig{
			TTLDays:   ttl,
			MaxPerDay: maxPerDay,
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			FilePath:   getEnv("LOG_FILE_PATH", ""),
			MaxSizeMB:  logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAgeDays: logMaxAge,
		},
		Redis: RedisConfig{
			Endpoint:      getEnv("COM_REDIS_ENDPOINT", "localhost"),
			Port:          getEnv("COM_REDIS_PORT", "6379"),
			Username:      getEnv("COM_REDIS_USERNAME", "myadmin"),
			Password:      getEnv("COM_REDIS_PASSWORD", "MyAdm1nP455w0rd"),
			Database:      getEnvAsInt("COM_REDIS_DATABASE", 0),
			PoolSize:      getEnvAsInt("COM_REDIS_POOL_SIZE", 10),
			MinConnection: getEnvAsInt("COM_REDIS_MIN_CONNECTION", 1),
			Timeout:       getEnvAsDuration("COM_REDIS_TIMEOUT", 10*time.Second),
			TLS:           getEnv("COM_REDIS_TLS_ENABLED", "false"),
		},
	}

	return cfg, nil
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
