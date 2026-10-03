package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) FoodQuery() generated.FoodQueryResolver { return &foodQueryResolver{r} }

type foodQueryResolver struct{ *Resolver }

func (r *foodQueryResolver) Food(ctx context.Context, obj *model.FoodQuery, id int) (*model.FoodFood, error) {
	return r.Clients.FoodService.Food(ctx, id)
}

func (r *foodQueryResolver) Foods(ctx context.Context, obj *model.FoodQuery) ([]*model.FoodFood, error) {
	return r.Clients.FoodService.Foods(ctx)
}

func (r *foodQueryResolver) SearchFoods(ctx context.Context, obj *model.FoodQuery, name string) ([]*model.FoodFood, error) {
	return r.Clients.FoodService.SearchFoods(ctx, name)
}
