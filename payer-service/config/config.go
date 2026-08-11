package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig    `json:"server"`
	Secret    SecretConfig    `json:"secret"`
	Nats      NatsConfig      `json:"nats"`
	Cassandra CassandraConfig `json:"cassandra"`
	Cache     CacheConfig     `json:"cache"`
}

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
}

type NatsConfig struct {
	Dns string `json:"dns"`
}

type CacheConfig struct {
	Dns string `json:"dns"`
}

type CassandraConfig struct {
	Dns      string `json:"dns"`
	Username string `json:"username"`
	Password string `json:"password"`
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
