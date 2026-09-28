package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	Secret   SecretConfig
	Constant ConstantConfig
	Cache    CacheConfig
	Keycloak KeycloakConfig
}

type ServerConfig struct {
	ServerPort string
	GRPCServer string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
	MysqlSSLMode  bool
	MysqlDriver   string
}

type SecretConfig struct {
	JwtSecretKey string
	AuthorizeKey string
}

type ConstantConfig struct {
	Rbac string
}
type KeycloakConfig struct {
	Host          string
	AdminUser     string
	AdminPassword string
	AdminRealm    string
	LegacyWildfly bool
}

type Consul struct {
	LockIndex int
	Key       int
	Flags     int
	Value     string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

type CacheConfig struct {
	Host string
}

func ReadConf(filename string) (*Config, error) {
	config := &Config{}
	buffer, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return config, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(buffer, config); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	return config, nil
}

func ReadFileEnv(dir string) (*Dir, error) {
	if err := godotenv.Load(dir); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return &Dir{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}, nil
}

func Path() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "config.yml"
}

func (c *Config) ApplyEnv() {
	setString(&c.Server.ServerPort, "SERVER_PORT")
	setString(&c.Secret.AuthorizeKey, "AUTHORIZE_KEY")
	setString(&c.Keycloak.Host, "KEYCLOAK_HOST")
	setString(&c.Keycloak.AdminUser, "KEYCLOAK_ADMIN_USER")
	setString(&c.Keycloak.AdminPassword, "KEYCLOAK_ADMIN_PASSWORD")
	setString(&c.Keycloak.AdminRealm, "KEYCLOAK_ADMIN_REALM")
	if v, err := strconv.ParseBool(os.Getenv("KEYCLOAK_LEGACY_WILDFLY")); err == nil {
		c.Keycloak.LegacyWildfly = v
	}
	if c.Server.ServerPort == "" {
		c.Server.ServerPort = "3031"
	}
}

func setString(target *string, key string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}
