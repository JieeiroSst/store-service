package userservice

import (
	"context"

	pb "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Module = fx.Options(
	fx.Provide(
		newConn,
		func(conn *grpc.ClientConn) pb.UserServiceClient { return pb.NewUserServiceClient(conn) },
		fx.Annotate(NewDirectory, fx.As(new(port.UserDirectory))),
	),
)

func newConn(lc fx.Lifecycle, cfg *config.Config) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(cfg.UserService.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return conn.Close() }})
	return conn, nil
}
