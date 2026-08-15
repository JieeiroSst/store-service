package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig      `json:"app"`
	HTTP     HTTPConfig     `json:"http"`
	Postgres PostgresConfig `json:"postgres"`
	Redis    RedisConfig    `json:"redis"`
	MinIO    MinIOConfig    `json:"minio"`
}

type AppConfig struct {
	Env      string `json:"env"`
	LogLevel string `json:"log_level"`
}

type HTTPConfig struct {
	Port            int      `json:"port"`
	ShutdownTimeout Duration `json:"shutdown_timeout"`
}

type PostgresConfig struct {
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	User           string   `json:"user"`
	Password       string   `json:"password"`
	DBName         string   `json:"dbname"`
	SSLMode        bool     `json:"sslmode"`
	MaxConns       int32    `json:"max_conns"`
	MigrationsPath string   `json:"migrations_path"`
	ConnectTimeout Duration `json:"connect_timeout"`
}

func (c PostgresConfig) DSN() string {
	sslmode := "disable"
	if c.SSLMode {
		sslmode = "require"
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     "/" + c.DBName,
		RawQuery: "sslmode=" + sslmode,
	}
	return u.String()
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type MinIOConfig struct {
	Endpoint        string   `json:"endpoint"`
	AccessKeyID     string   `json:"access_key_id"`
	SecretAccessKey string   `json:"secret_access_key"`
	UseSSL          bool     `json:"use_ssl"`
	Bucket          string   `json:"bucket"`
	PresignTTL      Duration `json:"presign_ttl"`
}

func (c *Config) applyDefaults() {
	if c.App.Env == "" {
		c.App.Env = "development"
	}
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.HTTP.Port == 0 {
		c.HTTP.Port = 8080
	}
	if c.HTTP.ShutdownTimeout == 0 {
		c.HTTP.ShutdownTimeout = Duration(10 * time.Second)
	}
	if c.Postgres.Host == "" {
		c.Postgres.Host = "localhost"
	}
	if c.Postgres.Port == 0 {
		c.Postgres.Port = 5432
	}
	if c.Postgres.User == "" {
		c.Postgres.User = "postgres"
	}
	if c.Postgres.Password == "" {
		c.Postgres.Password = "postgres"
	}
	if c.Postgres.DBName == "" {
		c.Postgres.DBName = "photo_service"
	}
	if c.Postgres.MaxConns == 0 {
		c.Postgres.MaxConns = 10
	}
	if c.Postgres.MigrationsPath == "" {
		c.Postgres.MigrationsPath = "migrations"
	}
	if c.Postgres.ConnectTimeout == 0 {
		c.Postgres.ConnectTimeout = Duration(5 * time.Second)
	}
	if c.Redis.Addr == "" {
		c.Redis.Addr = "localhost:6379"
	}
	if c.MinIO.Endpoint == "" {
		c.MinIO.Endpoint = "localhost:9000"
	}
	if c.MinIO.AccessKeyID == "" {
		c.MinIO.AccessKeyID = "minioadmin"
	}
	if c.MinIO.SecretAccessKey == "" {
		c.MinIO.SecretAccessKey = "minioadmin"
	}
	if c.MinIO.Bucket == "" {
		c.MinIO.Bucket = "photo-service"
	}
	if c.MinIO.PresignTTL == 0 {
		c.MinIO.PresignTTL = Duration(time.Hour)
	}
}

type bootstrap struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func readBootstrap(dir string) (*bootstrap, error) {
	_ = godotenv.Load(dir)

	b := &bootstrap{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}
	if b.HostConsul == "" || b.KeyConsul == "" || b.ServiceConsul == "" {
		return nil, fmt.Errorf("missing HostConsul/KeyConsul/ServiceConsul in %s or the environment", dir)
	}
	return b, nil
}

func Load() (*Config, error) {
	b, err := readBootstrap(".env")
	if err != nil {
		return nil, err
	}
	return loadFromConsul(b.HostConsul, b.KeyConsul, b.ServiceConsul)
}
