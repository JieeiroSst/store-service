package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Secret    SecretConfig
	Nats      NatsConfig
	Cassandra CassandraConfig
	Cache     CacheConfig
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

type CacheConfig struct {
	Dns string
}

type CassandraConfig struct {
	Dns      string
	Username string
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
