package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/config"
	"github.com/JIeeiroSst/shortlink-service/internal/repository"
	"github.com/JIeeiroSst/shortlink-service/internal/usecase"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	grpcService "github.com/JIeeiroSst/lib-gateway/shortlink-service/gateway/shortlink-service"
	serverHttp "github.com/JIeeiroSst/shortlink-service/internal/delivery/http"
)

var (
	ecosystem = ".env"
)

func runAPIV1() {
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "shortlink-service",
	})

	ctx := context.Background()
	mux := runtime.NewServeMux()

	config, _ := config.InitializeConfiguration(ecosystem)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Server.PortGrpcServer))
	if err != nil {
		logger.WithContext(ctx).Error("failed to listen", zap.Error(err))
	}

	db := postgres.NewPostgresConn(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  config.Postgres.PostgresqlSSLMode,
	})

	repository := repository.NewRepositories(db)
	cahce := expire.NewCacheHelper(config.Cache.Host)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:  repository,
		Expire: cahce,
		Domain: config.Host.Domain,
	})

	server := serverHttp.NewHandlerV1(usecase)
	grpcServer := grpc.NewServer()

	grpcService.RegisterShortlinkServiceServer(grpcServer, server)

	go func() {
		log.Printf("Starting gRPC server on :%v", config.Server.PortGrpcServer)
		logger.WithContext(ctx).Error("Starting gRPC server on", zap.String("PortGrpcServer", config.Server.PortGrpcServer))
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithContext(ctx).Error("Starting gRPC server", zap.Error(err))
		}
	}()

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = grpcService.RegisterShortlinkServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", config.Server.PortGrpcServer), opts)
	if err != nil {
		logger.WithContext(ctx).Error("RegisterAuthorizeServiceHandlerFromEndpoint", zap.Error(err))
	}

	log.Printf("Starting HTTP server on :%v", config.Server.PortHttpServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", config.Server.PortHttpServer), mux); err != nil {
		logger.WithContext(ctx).Debug("ListenAndServe", zap.Error(err))
	}
}

func runAPIV2() {
	r := gin.Default()
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "shortlink-service",
	})

	config, _ := config.InitializeConfiguration(ecosystem)

	db := postgres.NewPostgresConn(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  config.Postgres.PostgresqlSSLMode,
	})

	repository := repository.NewRepositories(db)
	cahce := expire.NewCacheHelper(config.Cache.Host)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:  repository,
		Expire: cahce,
		Domain: config.Host.Domain,
	})

	serverHttp.NewRouter(r, usecase)

	r.Run(config.Server.PortGinServer)
}
