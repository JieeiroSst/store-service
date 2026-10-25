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

	"github.com/JIeeiroSst/media-service/config"
	httpV1 "github.com/JIeeiroSst/media-service/internal/delivery/http"
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/media-service/internal/usecase"
	"github.com/JIeeiroSst/media-service/middleware"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/elastic"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/gin-gonic/gin"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	router := gin.Default()
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Error(context.Background(), "error %v", err)
	}
	consul := consul.NewConfigConsul(dirEnv.HostConsul,
		dirEnv.KeyConsul, dirEnv.ServiceConsul)
	var config config.Config
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Error(context.Background(), "error %v", err)
	}

	db := openPostgres(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  true,
	})

	cache := expire.NewCacheHelper(config.Cache.Dns)

	db.AutoMigrate()

	if err := json.Unmarshal(conf, &config); err != nil {
		logger.Error(context.Background(), "error %v", err)
	}

	elastic := elastic.NewElastic(config.Elastic.DNS)

	repository := repository.NewRepositories(db, elastic)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:       repository,
		CacheHelper: cache,
	})
	middleware := middleware.Newmiddleware(config.Secret.JwtSecretKey)

	httpServer := httpV1.NewHandler(usecase, middleware)

	httpServer.Init(router)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Server.PortServer),
		Handler: router,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Info(context.Background(), "listen: %s\n", err)
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

func openPostgres(c postgres.PostgresConfig) *gorm.DB {
	sslMode := "disable"
	if c.PostgresqlSSLMode {
		sslMode = "require"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.PostgresqlHost, c.PostgresqlUser, c.PostgresqlPassword, c.PostgresqlDbname, c.PostgresqlPort, sslMode)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	return db
}
