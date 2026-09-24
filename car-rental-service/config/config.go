package config

import (
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	Cache    CacheConfig    `json:"cache"`
	Secret   SecretConfig   `json:"secret"`
}

type ServerConfig struct {
	PortHttpServer string `json:"port_http_server"`
	PortGrpcServer string `json:"port_grpc_server"`
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

// SecretConfig holds the JWT settings shared with user-service, which issues the tokens.
type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
	AdminRole    string `json:"admin_role"`
	StaffRole    string `json:"staff_role"`
}

type CacheConfig struct {
	Host string `json:"host"`
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
