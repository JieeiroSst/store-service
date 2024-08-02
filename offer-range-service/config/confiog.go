package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Secret   SecretConfig
	RabbitMQ RabbitMQ
	Redis    Redis
	Postgres PostgresConfig
}

type ServerConfig struct {
	PortServer string
}

type Redis struct {
	Dns string
}

type RabbitMQ struct {
	Host           string
	Port           string
	User           string
	Password       string
	Exchange       string
	Queue          string
	RoutingKey     string
	ConsumerTag    string
	WorkerPoolSize int
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

type SecretConfig struct {
	JwtSecretKey string
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

func ReadConf(filename string) (*Config, error) {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		fmt.Printf("err: %v\n", err)

	}
	return config, nil
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
