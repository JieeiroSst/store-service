package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/payer-service/config"
	"github.com/JIeeiroSst/payer-service/internal/delivery/consumer"
	httpV1 "github.com/JIeeiroSst/payer-service/internal/delivery/http"
	"github.com/JIeeiroSst/payer-service/internal/repository"
	"github.com/JIeeiroSst/payer-service/internal/usecase"
	"github.com/JIeeiroSst/payer-service/middleware"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/cassandra"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/nats"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Error(context.Background(), err.Error())
	}
	consul := consul.NewConfigConsul(dirEnv.HostConsul,
		dirEnv.KeyConsul, dirEnv.ServiceConsul)
	var config config.Config
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Error(context.Background(), err.Error())
	}

	if err := json.Unmarshal(conf, &config); err != nil {
		logger.Error(context.Background(), err.Error())
	}

	db := cassandra.NewCassandra(cassandra.Cassandra{
		Dns:      config.Cassandra.Dns,
		Password: config.Cassandra.Password,
		Username: config.Cassandra.Username,
	})

	repository := repository.NewRepositories(db)
	cache := expire.NewCacheHelper(config.Cache.Dns)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:       repository,
		CacheHelper: cache,
	})
	nats := nats.ConnectNats(config.Nats.Dns)

	middleware := middleware.Newmiddleware(config.Secret.JwtSecretKey)

	httpServer := httpV1.NewHandler(usecase, middleware)

	cr := consumer.NewConsumer(usecase, nats)

	httpServer.Init(router)
	cr.Start(context.Background())

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Server.PortServer),
		Handler: router,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(context.Background(), "Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Info(context.Background(), "Server Shutdown: %v", err)
	}
	select {
	case <-ctx.Done():
		logger.Info(context.Background(), "timeout of 5 seconds.")
	}
	logger.Info(context.Background(), "Server exiting")
}
