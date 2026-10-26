package authorize_proxy

import (
	"context"
	"time"

	permissionpb "github.com/JIeeiroSst/gateway-api/proxy/pb/permissionpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthorizeProxy interface {
	CheckPermission(ctx context.Context, userID, object, action string) (bool, error)
	Close() error
}

type authorizeProxy struct {
	conn    *grpc.ClientConn
	api     permissionpb.PermissionServiceClient
	timeout time.Duration
}

func NewAuthorizeProxy(address string, timeout time.Duration, opts ...grpc.DialOption) (AuthorizeProxy, error) {
	conn, err := grpc.NewClient(address, append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, opts...)...)
	if err != nil {
		return nil, err
	}
	return &authorizeProxy{
		conn:    conn,
		api:     permissionpb.NewPermissionServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (p *authorizeProxy) CheckPermission(ctx context.Context, userID, object, action string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	res, err := p.api.CheckPermission(ctx, &permissionpb.CheckPermissionRequest{
		UserId: userID,
		Object: object,
		Action: action,
	})
	if err != nil {
		return false, err
	}
	return res.GetAllowed(), nil
}

func (p *authorizeProxy) Close() error { return p.conn.Close() }
