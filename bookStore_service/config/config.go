package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Secret   SecretConfig
	Constant ConstantConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	ServerPort string
	GRPCServer string
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

type SecretConfig struct {
	AuthorizeKey string
}

type ConstantConfig struct {
	Rbac string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

type CacheConfig struct {
	Host string
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
