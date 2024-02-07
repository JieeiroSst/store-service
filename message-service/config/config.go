package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Cache  CacheConfig
	Mysql  MysqlConfig
	Secret SecretConfig
	Kafka  KafkaConfig
}

type ServerConfig struct {
	ServerPort string
	GRPCServer string
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

type Kafka struct {
	Server string
}

type CacheConfig struct {
	Host string
}

type KafkaConfig struct {
	KafkaURL string
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
