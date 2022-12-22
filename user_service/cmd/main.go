package main

import (
	"os"
	"strings"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/app"
	"github.com/JIeeiroSst/user-service/pkg/log"
	"github.com/gin-gonic/gin"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	router := gin.Default()

	nodeEnv := os.Getenv("production")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err.Error())
	}
	if !strings.EqualFold(nodeEnv, "") {
		conf, err = config.ReadFileConsul(dirEnv.ConsulDir)
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		dir := "config.yml"
		conf, err = config.ReadConf(dir)
		if err != nil {
			log.Error(err.Error())
		}
	}

	app := app.NewApp(conf)

	go func() {
		app.NewServerGrpc()
	}()

	app.NewUserApp(router)
}
