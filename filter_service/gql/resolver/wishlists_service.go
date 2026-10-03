package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) WishlistQuery() generated.WishlistQueryResolver { return &wishlistQueryResolver{r} }

type wishlistQueryResolver struct{ *Resolver }

func (r *wishlistQueryResolver) Greet(ctx context.Context, obj *model.WishlistQuery) (*string, error) {
	return r.Clients.WishlistsService.Greet(ctx)
}

func (r *wishlistQueryResolver) Customers(ctx context.Context, obj *model.WishlistQuery) ([]*model.WishlistCustomer, error) {
	return r.Clients.WishlistsService.Customers(ctx)
}
