package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/cdn-service/internal/repository"
	"github.com/JIeeiroSst/cdn-service/internal/usecase"
	"github.com/JIeeiroSst/cdn-service/pkg/db"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	serverHttp "github.com/JIeeiroSst/cdn-service/internal/delivery/http"
	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"
)

var (
	ecosystem = ".env"
)

func initializeLogger() {
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "cdn-service",
	})
}

func initializeConfig() *config.Config {
	config, err := config.InitializeConfiguration(ecosystem)
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}
	return config
}

func initializeDatabase(config *config.Config) *db.DBInstance {
	connectionString := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v",
		config.Postgres.PostgresqlHost,
		config.Postgres.PostgresqlPort,
		config.Postgres.PostgresqlUser,
		config.Postgres.PostgresqlPassword,
		config.Postgres.PostgresqlDbname,
		config.Postgres.PostgresqlSSLMode,
	)
	return db.GetInstance(connectionString)
}

func initializeUsecase(config *config.Config, dbInstance *db.DBInstance) *usecase.Usecase {
	repository := repository.NewRepositories(dbInstance.DB)
	cache := expire.NewCacheHelper(config.Cache.Host)
	return usecase.NewUsecase(usecase.Dependency{
		Repos:    repository,
		Cache:    cache,
		BaseHost: config.BaseHost,
	})
}

func runGRPCServer(ctx context.Context, config *config.Config, server pb.FileServiceServer) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Server.PortGrpcServer))
	if err != nil {
		logger.WithContext(ctx).Error("failed to listen", zap.Error(err))
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, server)

	go func() {
		log.Printf("Starting gRPC server on :%v", config.Server.PortGrpcServer)
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithContext(ctx).Error("Starting gRPC server", zap.Error(err))
		}
	}()
}

func runHTTPServer(ctx context.Context, config *config.Config, mux *runtime.ServeMux) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterFileServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", config.Server.PortGrpcServer), opts)
	if err != nil {
		logger.WithContext(ctx).Error("RegisterFileServiceHandlerFromEndpoint", zap.Error(err))
		return
	}

	log.Printf("Starting HTTP server on :%v", config.Server.PortHttpServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", config.Server.PortHttpServer), mux); err != nil {
		logger.WithContext(ctx).Debug("ListenAndServe", zap.Error(err))
	}
}

func runAPIV1() {
	initializeLogger()
	ctx := context.Background()
	config := initializeConfig()
	dbInstance := initializeDatabase(config)
	usecase := initializeUsecase(config, dbInstance)

	server := serverHttp.NewHandlerV1(usecase)
	mux := runtime.NewServeMux()

	runGRPCServer(ctx, config, server)
	runHTTPServer(ctx, config, mux)
}

func runAPIV2() {
	initializeLogger()
	config := initializeConfig()
	dbInstance := initializeDatabase(config)
	usecase := initializeUsecase(config, dbInstance)

	r := gin.Default()
	serverHttp.NewRouter(r, usecase, config.BaseHost)

	log.Printf("Starting Gin HTTP server on :%v", config.Server.PortGinServer)
	if err := r.Run(fmt.Sprintf(":%v", config.Server.PortGinServer)); err != nil {
		log.Fatalf("Failed to start Gin server: %v", err)
	}
}
