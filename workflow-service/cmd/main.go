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

	"github.com/JIeeiroSst/workflow-service/config"
	"github.com/JIeeiroSst/workflow-service/internal/activities/card"
	httpServer "github.com/JIeeiroSst/workflow-service/internal/delivery/http"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/internal/usecase"
	"github.com/JIeeiroSst/workflow-service/pkg/consul"
	"github.com/JIeeiroSst/workflow-service/pkg/log"
	"github.com/JIeeiroSst/workflow-service/pkg/postgres"
	"github.com/JIeeiroSst/workflow-service/pkg/temporal"
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
		log.Error("", err)
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Error("", err)
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Error("", err)
		}
	}

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", conf.Postgres.PostgresqlHost,
		conf.Postgres.PostgresqlPort, conf.Postgres.PostgresqlUser, conf.Postgres.PostgresqlPassword, conf.Postgres.PostgresqlDbname)

	database := postgres.NewDatabase(psqlconn)
	repository := repository.NewRepositories(database.DB)
	temporal := temporal.NewWorkflow(conf.Temporal.Host)
	card := card.NewCardWorkflow(repository)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Temporal: temporal,
		Card:     card,
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
		log.Error("Server Shutdown: %v", err)
	}
	select {
	case <-ctx.Done():
		log.Info("timeout of 5 seconds.")
	}
	log.Info("Server exiting")
}
