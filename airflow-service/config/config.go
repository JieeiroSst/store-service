package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig
	Airflow AirflowConfig
}

type ServerConfig struct {
	ServerPort string
}

type AirflowConfig struct {
	Host     string
	Scheme   string
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
