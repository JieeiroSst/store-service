package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/grpcerr"
	"google.golang.org/grpc"
)

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func derefInt32(p *int) int32 { return int32(deref(p)) }

func int32Ptr(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}

func ptr[T any](v T) *T { return &v }

func grpcCall[T any](ctx context.Context, p *proxy.Proxy, service string, call func(context.Context, *grpc.ClientConn) (T, error)) (T, error) {
	conn, callCtx, cancel := p.GRPC(ctx, service)
	defer cancel()
	res, err := call(callCtx, conn)
	if err != nil {
		var zero T
		return zero, grpcerr.From(ctx, service, err)
	}
	return res, nil
}

func currentUser(ctx context.Context) (auth.User, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return auth.User{}, auth.NewError(ctx, auth.CodeUnauthenticated, "authentication required")
	}
	return user, nil
}
