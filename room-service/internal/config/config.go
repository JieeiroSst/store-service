// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/JIeeiroSst/room-service/internal/core/domain/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	// Server settings
	ServerPort int
	ServerHost string

	// Database settings
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// JWT settings
	JWTSecret     string
	JWTExpiryTime time.Duration

	// WebSocket settings
	WSReadBufferSize  int
	WSWriteBufferSize int

	// CORS settings
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string

	// Rate limiting
	RateLimit     int
	RateLimitTime time.Duration
}

// NewConfig creates a new configuration instance with default values
// and overrides them with environment variables if they exist
func NewConfig() *Config {
	config := &Config{
		// Server defaults
		ServerPort: getEnvInt("SERVER_PORT", 8081),
		ServerHost: getEnvStr("SERVER_HOST", "0.0.0.0"),

		// Database defaults
		DBHost:     getEnvStr("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 3306),
		DBUser:     getEnvStr("DB_USER", "chatuser"),
		DBPassword: getEnvStr("DB_PASSWORD", "chatpass123"),
		DBName:     getEnvStr("DB_NAME", "chatdb"),

		// JWT defaults
		JWTSecret:     getEnvStr("JWT_SECRET", "JWT_SECRET_KEY"),
		JWTExpiryTime: time.Duration(getEnvInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,

		// WebSocket defaults
		WSReadBufferSize:  getEnvInt("WS_READ_BUFFER_SIZE", 1024),
		WSWriteBufferSize: getEnvInt("WS_WRITE_BUFFER_SIZE", 1024),

		// CORS defaults
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods: getEnvSlice("ALLOWED_METHODS", []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		}),
		AllowedHeaders: getEnvSlice("ALLOWED_HEADERS", []string{
			"Origin", "Content-Type", "Accept", "Authorization",
		}),

		// Rate limiting defaults
		RateLimit:     getEnvInt("RATE_LIMIT", 100),
		RateLimitTime: time.Duration(getEnvInt("RATE_LIMIT_MINUTES", 1)) * time.Minute,
	}

	return config
}

// NewDB creates a new database connection using the configuration
func NewDB(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate the schemas
	if err := db.AutoMigrate(&models.Room{}, &models.Message{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Helper functions for environment variables

func getEnvStr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return append(defaultValue, value)
	}
	return defaultValue
}

// GetDSN returns the formatted database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// GetServerAddress returns the formatted server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

// GetJWTConfig returns JWT configuration settings
func (c *Config) GetJWTConfig() (string, time.Duration) {
	return c.JWTSecret, c.JWTExpiryTime
}

// GetWSBufferSizes returns WebSocket buffer sizes
func (c *Config) GetWSBufferSizes() (int, int) {
	return c.WSReadBufferSize, c.WSWriteBufferSize
}

// GetCORSConfig returns CORS configuration
func (c *Config) GetCORSConfig() ([]string, []string, []string) {
	return c.AllowedOrigins, c.AllowedMethods, c.AllowedHeaders
}

// GetRateLimitConfig returns rate limiting configuration
func (c *Config) GetRateLimitConfig() (int, time.Duration) {
	return c.RateLimit, c.RateLimitTime
}