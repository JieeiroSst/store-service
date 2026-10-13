package auth

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	CodeUnauthenticated    = "UNAUTHENTICATED"
	CodeForbidden          = "FORBIDDEN"
	CodeBadRequest         = "BAD_REQUEST"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeTooManyRequests    = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

func NewError(ctx context.Context, code, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Path:       graphql.GetPath(ctx),
		Message:    message,
		Extensions: map[string]interface{}{"code": code},
	}
}
