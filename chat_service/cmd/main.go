package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JIeeiroSst/chat-service/config"
	httpServer "github.com/JIeeiroSst/chat-service/internal/delivery/http"
	"github.com/JIeeiroSst/chat-service/internal/repository"
	"github.com/JIeeiroSst/chat-service/internal/usecase"
	"github.com/JIeeiroSst/chat-service/pkg/cache"
	"github.com/JIeeiroSst/chat-service/pkg/consul"
	"github.com/JIeeiroSst/chat-service/pkg/log"
	"github.com/JIeeiroSst/chat-service/pkg/mongo"
	"github.com/JIeeiroSst/chat-service/pkg/snowflake"
	"github.com/go-chi/chi/v5"
	"github.com/JIeeiroSst/chat-service/internal/delivery/websocket"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	router := chi.NewRouter()
	nodeEnv := os.Getenv("NODE_ENV")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err)
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Error(err)
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Error(err)
		}
	}

	var cache = cache.NewCacheHelper(conf.Cache.Host)
	var snowflakeData = snowflake.NewSnowflake()

	client := mongo.NewMongo(conf.Mongo.Host)
	repository := repository.NewRepositories(client.Client)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repo:        *repository,
		CacheHelper: cache,
		Snowflake:   snowflakeData,
	})

	httpServer := httpServer.NewHttp(*usecase)
	httpServer.Init(router)

	websocket:= websocket.NewWebSocket(*usecase)
	websocket.SetupRoutes()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.ServerPort),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info(fmt.Sprintf("listen: %s\n", err))
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("Server Shutdown: %v", err))
	}
	select {
	case <-ctx.Done():
		log.Info("timeout of 5 seconds.")
	}
	log.Info("Server exiting")
}
