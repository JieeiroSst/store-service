// Package server wires together the gRPC server and the HTTP/JSON gateway,
// and registers their start/stop hooks with the fx lifecycle.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	authorizeGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/config"
	grpchandler "github.com/JieeiroSst/authorize-service/internal/adapter/primary/grpc"
	"github.com/JieeiroSst/authorize-service/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// Params groups everything the server needs, resolved by fx.
type Params struct {
	fx.In

	LC      fx.Lifecycle
	Cfg     *config.Config
	Handler *grpchandler.Handler
}

// New registers gRPC and HTTP-gateway servers with the fx lifecycle.
// Servers start on LC.OnStart and shut down gracefully on LC.OnStop.
func New(p Params) {
	grpcAddr := fmt.Sprintf(":%s", p.Cfg.Server.PortGrpcServer)
	httpAddr := fmt.Sprintf(":%s", p.Cfg.Server.PortHttpServer)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GrpcInterceptor()),
	)
	authorizeGrpc.RegisterAuthorizeServiceServer(grpcServer, p.Handler)
	reflection.Register(grpcServer)

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lg := logger.WithContext(ctx)

			// ── gRPC ──────────────────────────────────────────────────────
			lis, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				return fmt.Errorf("server: listen gRPC %s: %w", grpcAddr, err)
			}
			go func() {
				lg.Info("gRPC server listening", zap.String("addr", grpcAddr))
				if err := grpcServer.Serve(lis); err != nil {
					lg.Error("gRPC server stopped", zap.Error(err))
				}
			}()

			// ── HTTP gateway ──────────────────────────────────────────────
			mux := runtime.NewServeMux()
			opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
			if err := authorizeGrpc.RegisterAuthorizeServiceHandlerFromEndpoint(
				ctx, mux, grpcAddr, opts,
			); err != nil {
				return fmt.Errorf("server: register gateway: %w", err)
			}
			httpServer := &http.Server{Addr: httpAddr, Handler: mux}
			go func() {
				lg.Info("HTTP gateway listening", zap.String("addr", httpAddr))
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					lg.Error("HTTP gateway stopped", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.WithContext(ctx).Info("server: graceful shutdown")
			grpcServer.GracefulStop()
			return nil
		},
	})
}
