package main

import (
	"github.com/JieeiroSst/authorize-service/internal/app"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	go func ()  {
		app.NewGRPCServer()
	}()

	app.NewApp(router)
}
