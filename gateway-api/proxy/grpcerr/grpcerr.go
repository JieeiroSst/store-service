package grpcerr

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func From(ctx context.Context, service string, err error) error {
	if err == nil {
		return nil
	}
	st, _ := status.FromError(err)
	code, message := mapCode(st.Code()), st.Message()
	if code == auth.CodeServiceUnavailable {
		logger.Error(ctx, "%s: %v", service, err)
		message = "service is unavailable"
	}
	ext := map[string]interface{}{"code": code, "service": service}
	for _, d := range st.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok && info.Reason != "" {
			ext["reason"] = info.Reason
		}
	}
	return &gqlerror.Error{Path: graphql.GetPath(ctx), Message: message, Extensions: ext}
}

func mapCode(c codes.Code) string {
	switch c {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return auth.CodeBadRequest
	case codes.Unauthenticated:
		return auth.CodeUnauthenticated
	case codes.PermissionDenied:
		return auth.CodeForbidden
	case codes.NotFound:
		return auth.CodeNotFound
	case codes.AlreadyExists, codes.Aborted:
		return auth.CodeConflict
	case codes.ResourceExhausted:
		return auth.CodeTooManyRequests
	}
	return auth.CodeServiceUnavailable
}
