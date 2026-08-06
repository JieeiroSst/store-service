package server

import (
	"context"
	"fmt"
	"net"

	"github.com/JIeeiroSst/cdn-service/config"
	grpcadapter "github.com/JIeeiroSst/cdn-service/internal/adapter/primary/grpc"
	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCParams struct {
	fx.In

	LC      fx.Lifecycle
	Handler *grpcadapter.Handler
	Config  *config.Config
}

func NewGRPCServer(p GRPCParams) {
	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, p.Handler)

	addr := fmt.Sprintf(":%s", p.Config.Server.PortGrpcServer)

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					logger.WithContext(ctx).Error("grpc server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
	})
}
