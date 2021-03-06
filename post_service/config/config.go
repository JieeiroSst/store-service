package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
)

type Config struct {
	Server          ServerConfig
	Mysql           MysqlConfig
	Secret 		    SecretService
	RabbitMQ        RabbitMQ
	Redis			Redis
	Email           Email
}

type ServerConfig struct {
	PortServer    string
	PprofPort     string
}

type Redis struct {
	Dns           string
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


type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
	MysqlSSLMode  bool
	MysqlDriver   string
}

type SecretService struct {
	JwtSecretKey string
}

type ElasticsearchConfig struct {
	Dns string
}

type Email struct {
	NameEmail     string
	PasswordEmail string
	Port          string
	Host          string
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