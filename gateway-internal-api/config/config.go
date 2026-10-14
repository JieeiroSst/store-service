package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

type Config struct {
	Server ServerConfig `json:"server"`
}

func (c *Config) ApplyDefaults() {
	if v := os.Getenv("PORT"); v != "" {
		c.Server.PortServer = v
	}
	if c.Server.PortServer == "" {
		c.Server.PortServer = "8081"
	}
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
