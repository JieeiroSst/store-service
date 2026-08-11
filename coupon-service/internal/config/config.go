package config

import (
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Secret   SecretConfig   `json:"secret"`
	Postgres PostgresConfig `json:"postgres"`
	Cache    CacheConfig    `json:"cache"`
	Nats     NatsConfig     `json:"nats"`
}

type ServerConfig struct {
	PortHttpServer string `json:"port_http_server"`
	PortGrpcServer string `json:"port_grpc_server"`
}

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
}

type CacheConfig struct {
	URL string `json:"url"`
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

type NatsConfig struct {
	DNS string `json:"dns"`
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
