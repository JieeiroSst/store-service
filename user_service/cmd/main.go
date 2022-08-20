package main

import (
	"log"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/app"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	dir := "config.yml"
	conf, err := config.ReadConf(dir)
	if err != nil {
		log.Fatal(err)
	}

	app := app.NewApp(conf)

	go func() {
		app.NewServerGrpc()
	}()

	app.NewUserApp(router)
}
