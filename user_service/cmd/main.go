package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/pb"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/internal/usecase"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/consul"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/log"
	"github.com/JIeeiroSst/user-service/pkg/postgres"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	grpcServer "github.com/JIeeiroSst/user-service/internal/delivery/grpc"
	httpServer "github.com/JIeeiroSst/user-service/internal/delivery/http"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	router := gin.Default()

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err.Error())
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err = consul.ConnectConfigConsul()
	if err != nil {
		log.Error(err.Error())
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		conf.Postgres.PostgresqlHost, conf.Postgres.PostgresqlUser, conf.Postgres.PostgresqlPassword,
		conf.Postgres.PostgresqlDbname, conf.Postgres.PostgresqlPort)

	postgresConn := postgres.NewPostgresConn(dsn)
	postgresConn.AutoMigrate(&model.Users{}, &model.Role{})

	snowflake := snowflake.NewSnowflake()
	hash := hash.NewHash()
	token := token.NewToken(conf)

	repository := repository.NewRepositories(postgresConn)
	cache := expire.NewCacheHelper(conf.Redis.Dns)

	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflake,
		Hash:      hash,
		Token:     token,
		Cache:     cache,
	})

	httpServer := httpServer.NewHandler(*usecase)

	httpServer.Init(router)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.PortServer),
		Handler: router,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	go func() {
		s := grpc.NewServer()
		srv := &grpcServer.GRPCServer{}
		srv.NewGRPCServer(*usecase)
		pb.RegisterAuthenticationServer(s, srv)
		l, err := net.Listen("tcp", conf.Server.PortServerGrpc)
		if err != nil {
			log.Error(err)
		}
		if err := s.Serve(l); err != nil {
			log.Error(err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("Server Shutdown: %v", err))
	}
	select {
	case <-ctx.Done():
		log.Info("timeout of 5 seconds.")
	}
	log.Info("Server exiting")
}
