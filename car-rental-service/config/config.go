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
	Roles    RolesConfig    `json:"roles"`

	UserService UserServiceConfig `json:"user_service"`
}

type UserServiceConfig struct {
	BaseURL        string `json:"base_url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
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

// RolesConfig maps authorize-service roles onto car-rental's admin and staff
// levels (defaults: "admin" and "operator"; "super_admin" is always admin).
type RolesConfig struct {
	AdminRole string `json:"admin_role"`
	StaffRole string `json:"staff_role"`
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
