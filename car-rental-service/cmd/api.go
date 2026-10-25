package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/car-rental-service/config"
	"github.com/JIeeiroSst/car-rental-service/internal/auth"
	"github.com/JIeeiroSst/car-rental-service/internal/repository"
	"github.com/JIeeiroSst/car-rental-service/internal/usecase"
	"github.com/JIeeiroSst/car-rental-service/internal/userservice"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	serverHttp "github.com/JIeeiroSst/car-rental-service/internal/delivery/http"
	pb "github.com/JIeeiroSst/lib-gateway/car-rental-servcie/gateway/car-rental-servcie"
)

var (
	ecosystem = ".env"
)

func runAPI() {
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "car-rental-service",
	})
	ctx := context.Background()
	mux := runtime.NewServeMux()
	config, _ := config.InitializeConfiguration(ecosystem)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Server.PortGrpcServer))
	if err != nil {
		logger.WithContext(ctx).Error("failed to listen", zap.Error(err))
	}

	db := openPostgres(postgres.PostgresConfig{
		PostgresqlHost:     config.Postgres.PostgresqlHost,
		PostgresqlPort:     config.Postgres.PostgresqlPort,
		PostgresqlUser:     config.Postgres.PostgresqlUser,
		PostgresqlPassword: config.Postgres.PostgresqlPassword,
		PostgresqlDbname:   config.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  config.Postgres.PostgresqlSSLMode,
	})

	users := userservice.New(config.UserService.BaseURL, time.Duration(config.UserService.TimeoutSeconds)*time.Second)

	repository := repository.NewRepositories(db)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos: repository,
		Users: userservice.Directory{C: users},
	})

	authn, err := auth.New(auth.Config{
		AdminRole: config.Roles.AdminRole,
		StaffRole: config.Roles.StaffRole,
	}, users, auth.MethodLevels)
	if err != nil {
		logger.WithContext(ctx).Error("init auth", zap.Error(err))
		return
	}

	server := serverHttp.NewHandler(usecase, authn)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authn.UnaryInterceptor()))
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

func openPostgres(c postgres.PostgresConfig) *gorm.DB {
	sslMode := "disable"
	if c.PostgresqlSSLMode {
		sslMode = "require"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.PostgresqlHost, c.PostgresqlUser, c.PostgresqlPassword, c.PostgresqlDbname, c.PostgresqlPort, sslMode)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	return db
}
