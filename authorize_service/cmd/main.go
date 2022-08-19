package main

import (
	"log"

	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/app"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	conf, err := config.ReadConf("config.yml")
	if err != nil {
		log.Println(err)
	}

	app := app.NewApp(conf)


	go func() {
		app.NewGRPCServer()
	}()

	app.NewServerApp(router)
}
