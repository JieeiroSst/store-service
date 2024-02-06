package config

import (
	"os"

	"github.com/JIeeiroSst/real-time-service/constant"
	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Serect SerectConfig
}

type ServerConfig struct {
	ServerPort string
}

type SerectConfig struct {
	Key string
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

func ReadFileEnv(dir string) (*Dir, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	data := &Dir{
		HostConsul:    os.Getenv(constant.HostConsul),
		KeyConsul:     os.Getenv(constant.KeyConsul),
		ServiceConsul: os.Getenv(constant.ServiceConsul),
	}
	return data, nil
}
