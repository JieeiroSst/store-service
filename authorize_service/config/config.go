package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	Secret   SecretService
	Constant Constant
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

type SecretService struct {
	JwtSecretKey string
}

type Constant struct {
	Rbac string
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
