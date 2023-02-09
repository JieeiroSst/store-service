package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/JieeiroSst/authorize-service/utils"
	"github.com/ghodss/yaml"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	Secret   SecretConfig
	Constant ConstantConfig
	Cache    CacheConfig
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

type Consul struct {
	LockIndex int
	Key       int
	Flags     int
	Value     string
}

type Dir struct {
	ConsulDir string
}

type CacheConfig struct {
	Host string
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

func ReadFileConsul(fileDir string) (*Config, error) {
	var (
		config Config
		consul Consul
	)
	resp, err := http.Get(fileDir)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &consul); err != nil {
		return nil, err
	}

	a := utils.DecodeByte(consul.Value)

	if err := json.Unmarshal(a, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ReadFileEnv(dir string) (*Dir, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	data := &Dir{
		ConsulDir: os.Getenv("consul"),
	}
	return data, nil
}
