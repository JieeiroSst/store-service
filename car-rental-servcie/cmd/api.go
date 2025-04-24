package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/car-rental-servcie/config"
	"github.com/JIeeiroSst/car-rental-servcie/internal/repository"
	"github.com/JIeeiroSst/car-rental-servcie/internal/usecase"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	serverHttp "github.com/JIeeiroSst/car-rental-servcie/internal/delivery/http"
	pb "github.com/JIeeiroSst/lib-gateway/car-rental-servcie/gateway/car-rental-servcie"
)

var (
	ecosystem = ".env"
)

func runAPI() {
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "car-rental-servcie",
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
	pb.RegisterVehicleRentalServiceServer(grpcServer, server)

	go func() {
		log.Printf("Starting gRPC server on :%v", config.Server.PortGrpcServer)
		logger.WithContext(ctx).Error("Starting gRPC server on", zap.String("PortGrpcServer", config.Server.PortGrpcServer))
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithContext(ctx).Error("Starting gRPC server", zap.Error(err))
		}
	}()

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterVehicleRentalServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", config.Server.PortGrpcServer), opts)
	if err != nil {
		logger.WithContext(ctx).Error("RegisterAuthorizeServiceHandlerFromEndpoint", zap.Error(err))
	}

	log.Printf("Starting HTTP server on :%v", config.Server.PortHttpServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", config.Server.PortHttpServer), mux); err != nil {
		logger.WithContext(ctx).Debug("ListenAndServe", zap.Error(err))
	}
}
