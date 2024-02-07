package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/bookStore-service/config"
	"github.com/JIeeiroSst/bookStore-service/pkg/consul"
	"github.com/JIeeiroSst/bookStore-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Logger().Error(err.Error())
	}
	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Logger().Error(err.Error())
	}



	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.ServerPort),
		Handler: router,
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger().Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger().Error("Server Shutdown:")
	}
	select {
	case <-ctx.Done():
		logger.Logger().Info("timeout of 5 seconds.")
	}
	logger.Logger().Info("Server exiting")
}
