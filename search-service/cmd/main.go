package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/search-service/config"
	httpDelivery "github.com/JIeeiroSst/search-service/internal/delivery/http"
	"github.com/JIeeiroSst/search-service/internal/repository"
	"github.com/JIeeiroSst/search-service/internal/service"
	"github.com/JIeeiroSst/search-service/middleware"
	"github.com/JIeeiroSst/search-service/pkg/consul"
	"github.com/JIeeiroSst/search-service/pkg/elasticsearch"
	"github.com/JIeeiroSst/search-service/pkg/logger"
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

	esv7Client, err := elasticsearch.NewElasticSearch(conf.Elasticsearch.DNS)
	if err != nil {
		logger.Logger().Error(err.Error())
	}
	middleware := middleware.Newmiddleware(conf.Secret.AuthorizeKey)
	repository := repository.NewRepositories(esv7Client)
	service := service.NewUsecase(service.Dependency{
		Repos: repository,
	})

	httpServer := httpDelivery.NewHandler(middleware, *service)
	httpServer.Init(router)

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
		logger.Logger().Errorf("Server Shutdown: %v", err)
	}
	select {
	case <-ctx.Done():
		logger.Logger().Info("timeout of 5 seconds.")
	}
	logger.Logger().Info("Server exiting")
}
