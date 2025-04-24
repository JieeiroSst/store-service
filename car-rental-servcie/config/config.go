package config

import (
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	PortHttpServer string
	PortGrpcServer string
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

type CacheConfig struct {
	Host string
}

func InitializeConfiguration(dir string) (*Config, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	consul := consul.NewConfigConsul(os.Getenv("HostConsul"), os.Getenv("KeyConsul"), os.Getenv("ServiceConsul"))
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(conf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
