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
	"github.com/JIeeiroSst/calculate-service/model"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.ConfigZap().Error(err.Error())
	}
	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	var config config.Config
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.ConfigZap().Error(err.Error())
	}

	if err := json.Unmarshal(conf, &config); err != nil {
		logger.ConfigZap().Error(err.Error())
	}

	db := postgres.NewPostgresConn(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  true,
	})

	db.AutoMigrate(&model.CampaignConfig{})

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
