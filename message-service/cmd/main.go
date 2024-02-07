package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/message-service/config"
	httpDelivery "github.com/JIeeiroSst/message-service/internal/delivery/http"
	"github.com/JIeeiroSst/message-service/internal/repository"
	"github.com/JIeeiroSst/message-service/internal/usecase"
	"github.com/JIeeiroSst/message-service/middleware"
	"github.com/JIeeiroSst/message-service/pkg/consul"
	"github.com/JIeeiroSst/message-service/pkg/kafka"
	"github.com/JIeeiroSst/message-service/pkg/logger"
	"github.com/JIeeiroSst/message-service/pkg/mysql"
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

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
	)
	mysqlOrm := mysql.NewMysqlConn(dns)
	queueKakfa := kafka.NetKafkaWriter(conf.Kafka.KafkaURL)

	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:      repository,
		QueueKakfa: queueKakfa,
	})

	middleware := middleware.Newmiddleware(conf.Secret.AuthorizeKey)

	httpServer := httpDelivery.NewHandler(middleware, *usecase)

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
		logger.Logger().Error("Server Shutdown:")
	}
	select {
	case <-ctx.Done():
		logger.Logger().Info("timeout of 5 seconds.")
	}
	logger.Logger().Info("Server exiting")
}
