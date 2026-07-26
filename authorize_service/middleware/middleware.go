package middleware

import (
	"context"
	"fmt"

	"github.com/JieeiroSst/authorize-service/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── gRPC server interceptor ──────────────────────────────────────────────────

func GrpcInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		log.Info(fmt.Sprintf("gRPC call: %s", info.FullMethod))

		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("gRPC panic on %s: %v", info.FullMethod, r))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
