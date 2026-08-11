package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig        `json:"server"`
	Secret        SecretConfig        `json:"secret"`
	Elasticsearch ElasticsearchConfig `json:"elasticsearch"`
}

type ServerConfig struct {
	ServerPort string `json:"server_port"`
}

type SecretConfig struct {
	AuthorizeKey string `json:"authorize_key"`
}

type ElasticsearchConfig struct {
	DNS string `json:"dns"`
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
