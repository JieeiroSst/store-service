package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/cdn-service/config"
	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GatewayParams struct {
	fx.In

	LC     fx.Lifecycle
	Config *config.Config
}

// NewGatewayServer exposes the gRPC FileService as REST/JSON by proxying to
// the gRPC server started by NewGRPCServer.
func NewGatewayServer(p GatewayParams) {
	mux := runtime.NewServeMux()
	grpcEndpoint := fmt.Sprintf("localhost:%s", p.Config.Server.PortGrpcServer)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.Config.Server.PortHttpServer),
		Handler: mux,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pb.RegisterFileServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts); err != nil {
				return err
			}
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.WithContext(ctx).Error("http gateway server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
