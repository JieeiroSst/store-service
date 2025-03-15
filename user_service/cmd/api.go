package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/internal/usecase"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	serverHttp "github.com/JIeeiroSst/user-service/internal/delivery/http"
)

var (
	ecosystem = ".env"
)

func runAPI() {
	config, _ := config.InitializeConfiguration(ecosystem)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Server.PortGrpcServer))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db := postgres.NewPostgresConn(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  config.Postgres.PostgresqlSSLMode,
	})

	db.AutoMigrate(&model.Users{}, &model.Role{})
	repository := repository.NewRepositories(db)
	cache := expire.NewCacheHelper(config.Redis.Dns)
	hash := hash.NewHash()
	token := token.NewToken(config.Secret.JwtSecretKey)

	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos: repository,
		Hash:  hash,
		Token: token,
		Cache: cache,
	})

	server := serverHttp.NewHandler(usecase)
	grpcServer := grpc.NewServer()
	userServiceGrpc.RegisterUserServiceServer(grpcServer, server)

	go func() {
		log.Printf("Starting gRPC server on :%v", config.Server.PortGrpcServer)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = userServiceGrpc.RegisterUserServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", config.Server.PortGrpcServer), opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	log.Printf("Starting HTTP server on :%v", config.Server.PortHttpServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", config.Server.PortHttpServer), mux); err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
}
