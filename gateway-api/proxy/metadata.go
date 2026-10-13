package proxy

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var unconfigured, _ = grpc.NewClient("passthrough:///unconfigured.invalid:1", grpc.WithTransportCredentials(insecure.NewCredentials()))

func outgoing(ctx context.Context) context.Context {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+user.Token, "x-user-id", user.ID)
}
