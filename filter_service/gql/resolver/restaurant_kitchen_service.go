package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RKitchenQuery() generated.RKitchenQueryResolver { return &rKitchenQueryResolver{r} }

type rKitchenQueryResolver struct{ *Resolver }

func (r *rKitchenQueryResolver) Categories(ctx context.Context, obj *model.RKitchenQuery) ([]*model.RKitchenCategory, error) {
	return r.Clients.RestaurantKitchenService.Categories(ctx)
}

func (r *rKitchenQueryResolver) Foods(ctx context.Context, obj *model.RKitchenQuery, limit *int, page *int, sort *string) (*model.RKitchenFoodPage, error) {
	return r.Clients.RestaurantKitchenService.Foods(ctx, limit, page, sort)
}

func (r *rKitchenQueryResolver) FoodsByIDs(ctx context.Context, obj *model.RKitchenQuery, ids string) ([]*model.RKitchenFood, error) {
	return r.Clients.RestaurantKitchenService.FoodsByIDs(ctx, ids)
}

func (r *rKitchenQueryResolver) Kitchens(ctx context.Context, obj *model.RKitchenQuery, limit *int, page *int, sort *string) (*model.RKitchenKitchenPage, error) {
	return r.Clients.RestaurantKitchenService.Kitchens(ctx, limit, page, sort)
}
