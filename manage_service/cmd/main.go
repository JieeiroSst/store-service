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

	"github.com/JIeeiroSst/manage-service/config"
	httpServer "github.com/JIeeiroSst/manage-service/internal/delivery/http"
	"github.com/JIeeiroSst/manage-service/internal/repository"
	"github.com/JIeeiroSst/manage-service/internal/usecase"
	"github.com/JIeeiroSst/manage-service/pkg/consul"
	"github.com/JIeeiroSst/manage-service/pkg/keycloak"
	"github.com/JIeeiroSst/manage-service/pkg/log"
	"github.com/go-chi/chi/v5"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	router := chi.NewRouter()

	nodeEnv := os.Getenv("NODE_ENV")

	log.Info("nodeEnv is " + nodeEnv)

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err.Error())
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Error(err.Error())
		}
	}

	keyclock := keycloak.NewKeycloak(conf.Keycloak.Host)
	repository := repository.NewRepositories(keyclock)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repo: repository,
	})

	httpServer := httpServer.NewHttp(usecase, conf)
	httpServer.Init(router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.ServerPort),
		Handler: router,
	}

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
