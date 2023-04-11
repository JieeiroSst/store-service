package main

import (
	"os"
	"strings"

	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/app"
	"github.com/JieeiroSst/authorize-service/pkg/consul"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/gin-gonic/gin"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	router := gin.Default()
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
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Error(err.Error())
		}
	}

	app := app.NewApp(conf)

	go func() {
		app.NewGRPCServer()
	}()

	app.NewServerApp(router)
}
