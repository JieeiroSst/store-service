package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Secret   SecretConfig   `json:"secret"`
	RabbitMQ RabbitMQ       `json:"rabbit_mq"`
	Redis    Redis          `json:"redis"`
	Postgres PostgresConfig `json:"postgres"`
}

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

type Redis struct {
	Dns string `json:"dns"`
}

type RabbitMQ struct {
	Host           string `json:"host"`
	Port           string `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	Exchange       string `json:"exchange"`
	Queue          string `json:"queue"`
	RoutingKey     string `json:"routing_key"`
	ConsumerTag    string `json:"consumer_tag"`
	WorkerPoolSize int    `json:"worker_pool_size"`
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

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
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
