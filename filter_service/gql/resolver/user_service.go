package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) UserQuery() generated.UserQueryResolver { return &userQueryResolver{r} }

type userQueryResolver struct{ *Resolver }

func (r *userQueryResolver) FindUser(ctx context.Context, obj *model.UserQuery, username *string, email *string, page *int, limit *int) (*model.UserFindUserResponse, error) {
	return r.Clients.UserService.FindUser(ctx, username, email, page, limit)
}

func (r *userQueryResolver) GetRole(ctx context.Context, obj *model.UserQuery, id int) (*model.UserGetRoleResponse, error) {
	return r.Clients.UserService.GetRole(ctx, id)
}

func (r *userQueryResolver) ListRoles(ctx context.Context, obj *model.UserQuery, limit *int, page *int) (*model.UserListRolesResponse, error) {
	return r.Clients.UserService.ListRoles(ctx, limit, page)
}
