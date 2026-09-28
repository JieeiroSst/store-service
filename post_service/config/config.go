package config

import (
	"os"
	"time"
)

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Auth   AuthConfig
	Minio  MinioConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

// AuthConfig points at user-service, which validates every bearer token.
type AuthConfig struct {
	UserServiceURL     string
	UserServiceTimeout time.Duration
}

type MinioConfig struct {
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "1236"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			Dbname:   getEnv("MYSQL_DBNAME", "post"),
		},
		Auth: AuthConfig{
			UserServiceURL:     getEnv("USER_SERVICE_URL", "http://user-service:1235"),
			UserServiceTimeout: getDuration("USER_SERVICE_TIMEOUT", 3*time.Second),
		},
		Minio: MinioConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:       getEnv("MINIO_ACCESS_KEY", ""),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", ""),
			BucketName:      getEnv("MINIO_BUCKET", "post-service"),
			UseSSL:          getEnv("MINIO_USE_SSL", "true") == "true",
		},
	}
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(key)); err == nil && d > 0 {
		return d
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
