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

	"github.com/JIeeiroSst/consumer-service/config"
	"github.com/JIeeiroSst/consumer-service/internal/delivery/consumer"
	httpServer "github.com/JIeeiroSst/consumer-service/internal/delivery/http"
	"github.com/JIeeiroSst/consumer-service/internal/model"
	"github.com/JIeeiroSst/consumer-service/internal/repository"
	"github.com/JIeeiroSst/consumer-service/internal/usecase"
	"github.com/JIeeiroSst/consumer-service/pkg/postgres"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.ConfigZap().Error(err.Error())
	}
	consul := logger.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	var config config.Config
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.ConfigZap().Error(err.Error())
	}

	if err := json.Unmarshal(conf, &config); err != nil {
		logger.ConfigZap().Error(err.Error())
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Postgres.PostgresqlHost, config.Postgres.PostgresqlUser, config.Postgres.PostgresqlPassword,
		config.Postgres.PostgresqlDbname, config.Postgres.PostgresqlPort)

	postgresConn := postgres.NewPostgresConn(dsn)
	postgresConn.AutoMigrate(&model.Consumer{})

	repository := repository.NewRepository(postgresConn)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos: repository,
	})
	nats := logger.ConnectNats(config.Nats.Dns)
	httpServer := httpServer.NewHandler(nats, usecase)
	consumer := consumer.NewConsumer(usecase, nats)

	httpServer.Init(router)
	consumer.Start(context.Background())

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Server.PortServer),
		Handler: router,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ConfigZap().Infof("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.ConfigZap().Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.ConfigZap().Error(fmt.Sprintf("Server Shutdown: %v", err))
	}
	select {
	case <-ctx.Done():
		logger.ConfigZap().Info("timeout of 5 seconds.")
	}
	logger.ConfigZap().Info("Server exiting")
}
