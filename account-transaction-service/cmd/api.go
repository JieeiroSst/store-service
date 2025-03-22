package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/Jieeirosst/account-transaction-service/config"
	serverHttp "github.com/Jieeirosst/account-transaction-service/internal/delivery/http"
	"github.com/Jieeirosst/account-transaction-service/internal/repository"
	"github.com/Jieeirosst/account-transaction-service/internal/usecase"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	grpService "github.com/JIeeiroSst/lib-gateway/account-transaction-service/gateway/account-transaction-service"
)

var (
	ecosystem = ".env"
)

func runAPI() {
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "authorize-service",
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
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos: repository,
	})

	server := serverHttp.NewHandler(usecase)
	grpcServer := grpc.NewServer()
	grpService.RegisterBankingServiceServer(grpcServer, server)

	go func() {
		log.Printf("Starting gRPC server on :%v", config.Server.PortGrpcServer)
		logger.WithContext(ctx).Error("Starting gRPC server on", zap.String("PortGrpcServer", config.Server.PortGrpcServer))
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithContext(ctx).Error("Starting gRPC server", zap.Error(err))
		}
	}()

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = grpService.RegisterBankingServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", config.Server.PortGrpcServer), opts)
	if err != nil {
		logger.WithContext(ctx).Error("RegisterAuthorizeServiceHandlerFromEndpoint", zap.Error(err))
	}

	log.Printf("Starting HTTP server on :%v", config.Server.PortHttpServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", config.Server.PortHttpServer), mux); err != nil {
		logger.WithContext(ctx).Debug("ListenAndServe", zap.Error(err))
	}
}
