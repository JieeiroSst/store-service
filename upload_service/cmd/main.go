package main

import (
	"os"
	"strings"

	"github.com/JIeeiroSst/upload-service/config"
	appServer "github.com/JIeeiroSst/upload-service/internal/app"
	"github.com/JIeeiroSst/upload-service/pkg/consul"
	"github.com/JIeeiroSst/upload-service/pkg/log"
	"github.com/gofiber/fiber/v2"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	app := fiber.New()

	nodeEnv := os.Getenv("NODE_ENV")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err.Error())
	}
	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		dir := ".env"
		conf, err = config.ConfigLocal(dir)
		if err != nil {
			log.Error(err.Error())
		}
	}

	appServer := appServer.NewServer(conf)

	appServer.NewServerApp(app)
}
