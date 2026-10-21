package config

import (
	"fmt"
	"net"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var Version = "dev"

type Config struct {
	Port          string
	OllamaBaseURL string
	OllamaModel   string
	MySQLDSN      string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config: failed to load .env: %w", err)
	}
	return FromEnv()
}

func FromEnv() (Config, error) {
	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		OllamaBaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11435"),
		OllamaModel:   os.Getenv("OLLAMA_MODEL"),
	}
	if cfg.OllamaModel == "" {
		return Config{}, fmt.Errorf("config: OLLAMA_MODEL is required")
	}

	dsn, err := mysqlDSN()
	if err != nil {
		return Config{}, err
	}
	cfg.MySQLDSN = dsn
	return cfg, nil
}

func mysqlDSN() (string, error) {
	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		return dsn, nil
	}

	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		return "", fmt.Errorf("config: MYSQL_DSN or MYSQL_HOST is required")
	}
	database := os.Getenv("MYSQL_DATABASE")
	if database == "" {
		return "", fmt.Errorf("config: MYSQL_DATABASE is required when MYSQL_HOST is set")
	}

	c := mysql.NewConfig()
	c.Net = "tcp"
	c.Addr = net.JoinHostPort(host, getEnv("MYSQL_PORT", "3306"))
	c.User = os.Getenv("MYSQL_USER")
	c.Passwd = os.Getenv("MYSQL_PASSWORD")
	c.DBName = database
	c.ParseTime = true
	return c.FormatDSN(), nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
