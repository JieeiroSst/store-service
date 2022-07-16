package main

import (
	"github.com/JIeeiroSst/user-service/internal/app"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	app.NewApp(router)
}
