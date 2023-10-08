package main

import (
	"github.com/JIeeiroSst/partner-service/internal/config"
	"github.com/JIeeiroSst/partner-service/internal/consul"
	"github.com/JIeeiroSst/partner-service/internal/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	logger.SetupLogger()
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Log.Error(err)
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Log.Error(err)
	}

	InitRoutes(conf)
}

func InitRoutes(conf *config.Config) {
	router := gin.Default()

	pprof.Register(router)

	err := router.Run(conf.Server.ServerPort)
	if err != nil {
		logger.Log.Error(err)
	}
}
