package main

import (
	"log"

	"github.com/JIeeiroSst/partner-service/internal/config"
	"github.com/JIeeiroSst/partner-service/internal/consul"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	InitRoutes(conf)
}

func InitRoutes(conf *config.Config) {
	router := gin.Default()

	pprof.Register(router)

	err := router.Run(conf.Server.ServerPort)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
