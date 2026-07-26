package di

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
)

func RegisterServer(lc fx.Lifecycle, cfg *config.Config, handler *grpcadapter.Handler) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Server.PortGrpcServer))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	userServiceGrpc.RegisterUserServiceServer(grpcServer, handler)

	mux := runtime.NewServeMux()
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Server.PortHttpServer),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Printf("Starting gRPC server on :%v", cfg.Server.PortGrpcServer)
				if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
					log.Printf("gRPC server stopped: %v", err)
				}
			}()

			opts := []grpc.DialOption{grpc.WithInsecure()}
			endpoint := fmt.Sprintf("localhost:%v", cfg.Server.PortGrpcServer)
			if err := userServiceGrpc.RegisterUserServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
				return fmt.Errorf("failed to register gateway: %w", err)
			}

			go func() {
				log.Printf("Starting HTTP server on :%v", cfg.Server.PortHttpServer)
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server stopped: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return httpServer.Shutdown(ctx)
		},
	})

	return nil
}
