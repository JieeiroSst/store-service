package config

import "os"

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	BaseHost BaseHostConfig
}

type ServerConfig struct {
	PortHttpServer string
	PortGrpcServer string
	PortGinServer  string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
}

type BaseHostConfig struct {
	DominServiceURL string
	BaseDirUpload   string
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "2234"),
			PortGrpcServer: getEnv("PORT_GRPC_SERVER", "2231"),
			PortGinServer:  getEnv("PORT_GIN_SERVER", "2232"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "cdn"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSLMODE", "disable") == "require",
		},
		BaseHost: BaseHostConfig{
			DominServiceURL: getEnv("DOMAIN_SERVICE_URL", "http://localhost:2232"),
			BaseDirUpload:   getEnv("BASE_DIR_UPLOAD", "./uploads"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
