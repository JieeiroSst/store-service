package graph

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const MaxRootFields = 30

func LimitRootFields(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx != nil && opCtx.Operation != nil {
		if n := len(graphql.CollectFields(opCtx, opCtx.Operation.SelectionSet, nil)); n > MaxRootFields {
			err := auth.NewError(ctx, auth.CodeBadRequest, fmt.Sprintf("operation selects %d root fields, the limit is %d", n, MaxRootFields))
			return graphql.OneShot(&graphql.Response{Errors: []*gqlerror.Error{err}})
		}
	}
	return next(ctx)
}
