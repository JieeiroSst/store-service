package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Secret   SecretConfig
	Nats     NatsConfig
	Postgres PostgresConfig
}

type ServerConfig struct {
	PortServer string
}

type SecretConfig struct {
	JwtSecretKey string
}

type NatsConfig struct {
	Dns string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
	PgDriver           string
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
