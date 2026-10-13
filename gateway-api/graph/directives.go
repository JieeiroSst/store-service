package graph

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/proxy/authorize_proxy"
	"github.com/JIeeiroSst/utils/logger"
)

func NewDirectiveRoot(authorize authorize_proxy.AuthorizeProxy) DirectiveRoot {
	return DirectiveRoot{
		Public: func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
			return next(ctx)
		},
		HasPermission: func(ctx context.Context, obj interface{}, next graphql.Resolver, object string, action string) (interface{}, error) {
			user, ok := auth.UserFromContext(ctx)
			if !ok {
				return nil, auth.NewError(ctx, auth.CodeUnauthenticated, "authentication required")
			}
			allowed, err := authorize.CheckPermission(ctx, user.ID, object, action)
			if err != nil {
				logger.Error(ctx, "check permission user=%s object=%s action=%s: %v", user.ID, object, action, err)
				return nil, auth.NewError(ctx, auth.CodeServiceUnavailable, "cannot verify permission, try again later")
			}
			if !allowed {
				return nil, auth.NewError(ctx, auth.CodeForbidden, "you do not have permission to perform this action")
			}
			return next(ctx)
		},
	}
}

func RequireAuthentication(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
	field := graphql.GetRootFieldContext(ctx)
	if field == nil || isPublic(field) {
		return next(ctx)
	}
	if _, ok := auth.UserFromContext(ctx); ok {
		return next(ctx)
	}
	graphql.AddError(ctx, auth.NewError(ctx, auth.CodeUnauthenticated, "authentication required"))
	return graphql.Null
}

func isPublic(field *graphql.RootFieldContext) bool {
	if strings.HasPrefix(field.Field.Name, "__") {
		return true
	}
	if field.Field.Definition == nil {
		return false
	}
	return field.Field.Definition.Directives.ForName("public") != nil
}
