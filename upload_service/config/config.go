package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port       string
	TokenImgBB string
	ImgBBApi   string
}

func Config() (*ServerConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}
	return &ServerConfig{
		Port:       os.Getenv("PORT"),
		TokenImgBB: os.Getenv("TOKEN_IMGBB"),
		ImgBBApi:   os.Getenv("IMGBB_API"),
	}, nil
}
