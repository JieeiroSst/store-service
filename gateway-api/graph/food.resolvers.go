package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) FoodCreateFood(ctx context.Context, input model.FoodFoodInput) (*model.FoodFood, error) {
	return restCall[*model.FoodFood](ctx, r.Proxy, "food-service", rest.Request{Method: "POST", Path: "/api/food", Body: input}, "")
}

func (r *mutationResolver) FoodUpdateFood(ctx context.Context, id int, input model.FoodFoodInput) (*model.FoodFood, error) {
	return restCall[*model.FoodFood](ctx, r.Proxy, "food-service", rest.Request{Method: "PUT", Path: rest.P("/api/food/%v", id), Body: input}, "")
}

func (r *mutationResolver) FoodDeleteFood(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "food-service", rest.Request{Method: "DELETE", Path: rest.P("/api/food/%v", id)}, "")
}

func (r *mutationResolver) FoodDeleteAllFoods(ctx context.Context) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "food-service", rest.Request{Method: "DELETE", Path: "/api/food/"}, "")
}

func (r *mutationResolver) FoodCreateAllergen(ctx context.Context, name string) (*model.FoodAllergen, error) {
	return restCall[*model.FoodAllergen](ctx, r.Proxy, "food-service", rest.Request{Method: "POST", Path: "/api/allergens", Body: map[string]any{"name": name}}, "")
}

func (r *mutationResolver) FoodCreateIngredient(ctx context.Context, name string, unit *string) (*model.FoodIngredient, error) {
	return restCall[*model.FoodIngredient](ctx, r.Proxy, "food-service", rest.Request{Method: "POST", Path: "/api/ingredient", Body: map[string]any{"name": name, "unit": unit}}, "")
}

func (r *queryResolver) Food(ctx context.Context, id int) (*model.FoodFood, error) {
	return restCall[*model.FoodFood](ctx, r.Proxy, "food-service", rest.Request{Method: "GET", Path: rest.P("/api/food/%v", id)}, "")
}

func (r *queryResolver) Foods(ctx context.Context) ([]*model.FoodFood, error) {
	return restCall[[]*model.FoodFood](ctx, r.Proxy, "food-service", rest.Request{Method: "GET", Path: "/api/food/"}, "")
}

func (r *queryResolver) FoodSearchFoods(ctx context.Context, name string) ([]*model.FoodFood, error) {
	return restCall[[]*model.FoodFood](ctx, r.Proxy, "food-service", rest.Request{Method: "GET", Path: "/api/food/search", Query: restQuery("name", name)}, "")
}
