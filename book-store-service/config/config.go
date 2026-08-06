package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Secret   SecretConfig
}

type ServerConfig struct {
	ServerPort string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
}

type SecretConfig struct {
	AuthorizeKey string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func ReadFileEnv(dir string) (*Dir, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	data := &Dir{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}
	return data, nil
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			ServerPort: getEnv("PORT", "8000"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "book_store"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSLMODE", "disable") == "require",
		},
		Secret: SecretConfig{
			AuthorizeKey: getEnv("AUTHORIZE_KEY", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
