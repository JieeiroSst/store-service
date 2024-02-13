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
	AuthorizeKey string
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
