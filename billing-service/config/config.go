package config

import "os"

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
	JwtSecretKey string
	AuthorizeKey string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

// FromEnv builds configuration straight from environment variables. Used
// when Consul is not configured (local dev, CI, Consul-less deployments).
func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			ServerPort: getEnv("PORT", "8000"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "timescaledb"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", "password"),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "billing"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSL_MODE", "false") == "true",
		},
		Secret: SecretConfig{
			JwtSecretKey: getEnv("JWT_SECRET_KEY", ""),
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
