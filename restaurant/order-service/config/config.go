package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Secret   SecretConfig   `json:"secret"`
	Redis    Redis          `json:"redis"`
	Postgres PostgresConfig `json:"postgres"`
	Nats     NatsConfig     `json:"nats"`
}

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresql_host"`
	PostgresqlPort     string `json:"postgresql_port"`
	PostgresqlUser     string `json:"postgresql_user"`
	PostgresqlPassword string `json:"postgresql_password"`
	PostgresqlDbname   string `json:"postgresql_dbname"`
	PostgresqlSSLMode  bool   `json:"postgresql_ssl_mode"`
	PgDriver           string `json:"pg_driver"`
}

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
}

type Redis struct {
	Dns string `json:"dns"`
}

type NatsConfig struct {
	Dns string `json:"dns"`
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
