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

	"github.com/JIeeiroSst/calculate-service/config"
	"github.com/JIeeiroSst/calculate-service/internal/delivery/consumer"
	http1 "github.com/JIeeiroSst/calculate-service/internal/delivery/http"
	"github.com/JIeeiroSst/calculate-service/internal/delivery/worker"
	"github.com/JIeeiroSst/calculate-service/internal/repository"
	"github.com/JIeeiroSst/calculate-service/internal/usecase"
	"github.com/JIeeiroSst/calculate-service/middleware"
	"github.com/JIeeiroSst/calculate-service/model"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/logger"
	nat "github.com/JIeeiroSst/utils/nats"
	"github.com/JIeeiroSst/utils/postgres"
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

	db := postgres.NewPostgresConn(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  true,
	})

	cache := expire.NewCacheHelper(config.Cache.DNS)
	middleware := middleware.Newmiddleware(config.Secret.JwtSecretKey)

	db.AutoMigrate(&model.CampaignTypeConfig{}, &model.CampaignConfig{})

	repository := repository.NewRepositories(db)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:       repository,
		CacheHelper: cache,
	})

	nats := nat.ConnectNats(config.Nats.Dns)

	httpServer := http1.NewHandler(usecase, middleware)
	worker := worker.NewWorker(usecase)
	consumer := consumer.NewConsumer(nats, usecase)

	worker.RunWorker()
	consumer.RunConsumer()
	httpServer.Init(router)

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
