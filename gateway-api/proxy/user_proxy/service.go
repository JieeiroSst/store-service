package user_proxy

import (
	"context"
	"time"

	userServiceGrpc "github.com/JIeeiroSst/gateway-api/proxy/pb/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserProxy interface {
	ValidateSession(ctx context.Context, token string) (Session, error)
	API() userServiceGrpc.UserServiceClient
	WithTimeout(ctx context.Context) (context.Context, context.CancelFunc)
	Close() error
}

type userProxy struct {
	conn    *grpc.ClientConn
	api     userServiceGrpc.UserServiceClient
	timeout time.Duration
}

func NewUserProxy(address string, timeout time.Duration) (UserProxy, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &userProxy{
		conn:    conn,
		api:     userServiceGrpc.NewUserServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (p *userProxy) ValidateSession(ctx context.Context, token string) (Session, error) {
	ctx, cancel := p.WithTimeout(ctx)
	defer cancel()

	res, err := p.api.ValidateSession(ctx, &userServiceGrpc.ValidateRequest{SessionToken: token})
	if err != nil {
		return Session{}, err
	}
	return Session{Valid: res.GetValid(), UserID: res.GetUserId()}, nil
}

func (p *userProxy) API() userServiceGrpc.UserServiceClient { return p.api }

func (p *userProxy) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, p.timeout)
}

func (p *userProxy) Close() error { return p.conn.Close() }
