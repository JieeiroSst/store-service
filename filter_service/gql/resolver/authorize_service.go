package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AuthzQuery() generated.AuthzQueryResolver { return &authzQueryResolver{r} }

type authzQueryResolver struct{ *Resolver }

func (r *authzQueryResolver) GetCasbinRules(ctx context.Context, obj *model.AuthzQuery, limit *int, page *int, sort *string, totalRows *string, totalPages *int) (*model.AuthzCasbinRuleList, error) {
	return r.Clients.AuthorizeService.GetCasbinRules(ctx, limit, page, sort, totalRows, totalPages)
}

func (r *authzQueryResolver) GetCasbinRuleByID(ctx context.Context, obj *model.AuthzQuery, id int) (*model.AuthzCasbinRule, error) {
	return r.Clients.AuthorizeService.GetCasbinRuleByID(ctx, id)
}
