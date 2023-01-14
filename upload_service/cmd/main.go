package main

import (
	"os"
	"strings"

	"github.com/JIeeiroSst/upload-service/common"
	"github.com/JIeeiroSst/upload-service/config"
	appServer "github.com/JIeeiroSst/upload-service/internal/app"
	"github.com/JIeeiroSst/upload-service/pkg/log"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var (
	cfg *config.ServerConfig
)

func main() {
	app := fiber.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	nodeEnv := os.Getenv(common.Production)
	if !strings.EqualFold(nodeEnv, "") {
		cfg, err = config.ConfigConsul()
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		cfg, err = config.ConfigLocal()
		if err != nil {
			log.Error(err.Error())
		}
	}

	appServer := appServer.NewServer(cfg)

	appServer.NewServerApp(app)
}
