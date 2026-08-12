package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	MinIO    MinIOConfig    `json:"minio"`
	Secret   SecretConfig   `json:"secret"`
	Edge     EdgeConfig     `json:"edge"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresql_host"`
	PostgresqlPort     string `json:"postgresql_port"`
	PostgresqlUser     string `json:"postgresql_user"`
	PostgresqlPassword string `json:"postgresql_password"`
	PostgresqlDbname   string `json:"postgresql_dbname"`
	PostgresqlSSLMode  bool   `json:"postgresql_ssl_mode"`
}

type MinIOConfig struct {
	Endpoint             string `json:"endpoint"`
	AccessKeyID          string `json:"access_key_id"`
	SecretAccessKey      string `json:"secret_access_key"`
	UseSSL               bool   `json:"use_ssl"`
	BucketName           string `json:"bucket_name"`
	MaxUploadSizeBytes   int64  `json:"max_upload_size_bytes"`
	PresignExpiryMinutes int    `json:"presign_expiry_minutes"`
}

func (m MinIOConfig) PresignExpiry() time.Duration {
	minutes := m.PresignExpiryMinutes
	if minutes <= 0 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}

type SecretConfig struct {
	APIKey string `json:"api_key"`
}

type EdgeConfig struct {
	BaseURL string `json:"base_url"`
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func ReadFileEnv(dir string) (*Dir, error) {
	if err := godotenv.Load(dir); err != nil {
		return nil, err
	}
	return &Dir{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}, nil
}
