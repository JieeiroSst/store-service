package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig
	Secret        SecretConfig
	Elasticsearch ElasticsearchConfig
}

type ServerConfig struct {
	ServerPort string
}

type SecretConfig struct {
	AuthorizeKey string
}

type ElasticsearchConfig struct {
	DNS string
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
