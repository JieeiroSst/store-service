package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port string
}

func Config() (*ServerConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}
	return &ServerConfig{
		Port: os.Getenv("PORT"),
	}, nil
}
