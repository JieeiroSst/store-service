package main

import (
	"fmt"
	"net/http"

	"github.com/JieeiroSst/gateway-service/config"
	"github.com/JieeiroSst/gateway-service/pkg/consul"
	"github.com/JieeiroSst/gateway-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.LoggerError(err.Error())
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.LoggerError(err.Error())
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.ServerPort),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LoggerError(err.Error())
		}
	}()
}
