package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	LogLevel string         `json:"log_level"`
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

func (c Config) DSN() string {
	sslmode := "disable"
	if c.Postgres.PostgresqlSSLMode {
		sslmode = "require"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Postgres.PostgresqlUser, c.Postgres.PostgresqlPassword,
		c.Postgres.PostgresqlHost, c.Postgres.PostgresqlPort,
		c.Postgres.PostgresqlDbname, sslmode)
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
