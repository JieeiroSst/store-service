package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BasketQuery() generated.BasketQueryResolver { return &basketQueryResolver{r} }

type basketQueryResolver struct{ *Resolver }

func (r *basketQueryResolver) Baskets(ctx context.Context, obj *model.BasketQuery) ([]*model.BasketBasket, error) {
	return r.Clients.BasketService.Baskets(ctx)
}

func (r *basketQueryResolver) Basket(ctx context.Context, obj *model.BasketQuery, id int) (*model.BasketBasket, error) {
	return r.Clients.BasketService.Basket(ctx, id)
}

func (r *basketQueryResolver) BasketLines(ctx context.Context, obj *model.BasketQuery) ([]*model.BasketBasketLine, error) {
	return r.Clients.BasketService.BasketLines(ctx)
}

func (r *basketQueryResolver) BasketLine(ctx context.Context, obj *model.BasketQuery, id int) (*model.BasketBasketLine, error) {
	return r.Clients.BasketService.BasketLine(ctx, id)
}

func (r *basketQueryResolver) BasketLineAttributes(ctx context.Context, obj *model.BasketQuery) ([]*model.BasketBasketLineAttribute, error) {
	return r.Clients.BasketService.BasketLineAttributes(ctx)
}

func (r *basketQueryResolver) BasketLineAttribute(ctx context.Context, obj *model.BasketQuery, id int) (*model.BasketBasketLineAttribute, error) {
	return r.Clients.BasketService.BasketLineAttribute(ctx, id)
}

func (r *basketQueryResolver) Orders(ctx context.Context, obj *model.BasketQuery) ([]*model.BasketOrder, error) {
	return r.Clients.BasketService.Orders(ctx)
}

func (r *basketQueryResolver) Order(ctx context.Context, obj *model.BasketQuery, id int) (*model.BasketOrder, error) {
	return r.Clients.BasketService.Order(ctx, id)
}

func (r *basketQueryResolver) User(ctx context.Context, obj *model.BasketQuery, id int) (*model.BasketUser, error) {
	return r.Clients.BasketService.User(ctx, id)
}
