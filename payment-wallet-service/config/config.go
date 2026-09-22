package config

import "os"

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8088"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "payment_wallet"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSLMODE", "disable") == "require",
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
