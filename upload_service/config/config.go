package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDns     string
	Port         string
	TokenImgBB   string
	ImgBBApi     string
	HostCacheDNS string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func ConfigLocal(dir string) (*Config, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}
	return &Config{
		Port:       os.Getenv("PORT"),
		TokenImgBB: os.Getenv("TOKEN_IMGBB"),
		ImgBBApi:   os.Getenv("IMGBB_API"),
		MongoDns:   os.Getenv("MONGO_DB"),
	}, nil
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
