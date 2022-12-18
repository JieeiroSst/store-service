package main

import (
	"log"
	"os"
	"strings"

	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/app"
	"github.com/gin-gonic/gin"
)

var (
	conf *config.Config
	err  error
)

func main() {
	router := gin.Default()
	nodeEnv := os.Getenv("production")
	if !strings.EqualFold(nodeEnv, "") {

	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Println(err)
		}
	}

	app := app.NewApp(conf)

	go func() {
		app.NewGRPCServer()
	}()

	app.NewServerApp(router)
}
