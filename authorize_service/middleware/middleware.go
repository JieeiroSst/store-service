package middleware

import (
	"context"
	"fmt"

	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── HTTP (Gin) middleware ────────────────────────────────────────────────────

type Middleware interface {
	AuthorizeControl() gin.HandlerFunc
}

type middleware struct {
	secret string
}

func NewMiddleware(secret string) Middleware {
	return &middleware{secret: secret}
}

func (m *middleware) AuthorizeControl() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.Request.Header.Get("Authorization")
		if ok := utils.DecodeBase(authorization, m.secret); !ok {
			log.Error(fmt.Sprintf("Unauthorized from %v", ctx.RemoteIP()))
			ctx.AbortWithStatusJSON(403, gin.H{"message": "Unauthorized"})
			return
		}
		ctx.Next()
	}
}

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
