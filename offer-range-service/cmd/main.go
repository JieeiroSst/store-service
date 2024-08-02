package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/offer-range-service/config"
	"github.com/JIeeiroSst/offer-range-service/pkg/consul"
	"github.com/JIeeiroSst/offer-range-service/pkg/log"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err.Error())
	}
	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		log.Error(err.Error())
	}

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.PortServer),
		Handler: router,
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("Server Shutdown: %v", err))
	}
	select {
	case <-ctx.Done():
		log.Info("timeout of 5 seconds.")
	}
	log.Info("Server exiting")
}
