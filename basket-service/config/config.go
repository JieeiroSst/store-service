package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Mysql  MysqlConfig
	Secret SecretConfig
	Cache  CacheConfig
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

type CacheConfig struct {
	Host     string
	Password string
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

// FromEnv builds configuration straight from environment variables. Used
// when Consul is not configured (local dev, CI, Consul-less deployments).
func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			ServerPort: getEnv("PORT", "8000"),
			GRPCServer: getEnv("GRPC_SERVER", ""),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "basket_service"),
			MysqlSSLMode:  getEnv("MYSQL_SSL_MODE", "false") == "true",
			MysqlDriver:   getEnv("MYSQL_DRIVER", "mysql"),
		},
		Secret: SecretConfig{
			JwtSecretKey: getEnv("JWT_SECRET_KEY", ""),
			AuthorizeKey: getEnv("AUTHORIZE_KEY", ""),
		},
		Cache: CacheConfig{
			Host:     getEnv("CACHE_HOST", "localhost:6379"),
			Password: getEnv("CACHE_PASSWORD", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
