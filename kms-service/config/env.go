package config

import (
	"log"

	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseDSN     string
	RedisAddr       string
	RedisPassword   string
	JWTSecret       string
	MasterKeyPath   string
	ServerPort      string
	KeyRotationDays int
	CacheExpiration time.Duration
	RateLimitPerMin int
	LogLevel        string
	HSMEnabled      bool
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env variables")
	}

	AppConfig = &Config{
		DatabaseDSN:     getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/kms_db?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		JWTSecret:       getEnv("JWT_SECRET", "your-jwt-secret-key"),
		MasterKeyPath:   getEnv("MASTER_KEY_PATH", "/etc/kms/master.key"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		KeyRotationDays: getEnvAsInt("KEY_ROTATION_DAYS", 365),
		CacheExpiration: time.Duration(getEnvAsInt("CACHE_EXPIRATION_MIN", 10)) * time.Minute,
		RateLimitPerMin: getEnvAsInt("RATE_LIMIT_PER_MIN", 100),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		HSMEnabled:      getEnvAsBool("HSM_ENABLED", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
