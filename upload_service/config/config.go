package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	MongoDns     string
	Port         string
	TokenImgBB   string
	ImgBBApi     string
	HostCacheDNS string
}

func ConfigLocal() (*ServerConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}
	return &ServerConfig{
		Port:       os.Getenv("PORT"),
		TokenImgBB: os.Getenv("TOKEN_IMGBB"),
		ImgBBApi:   os.Getenv("IMGBB_API"),
		MongoDns:   os.Getenv("MONGO_DB"),
	}, nil
}

func ConfigConsul() (*ServerConfig, error) {
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
