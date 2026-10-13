package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RKitchenCreateCategory(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "POST", Path: "/api/v1/category", Body: input}, "")
}

func (r *mutationResolver) RKitchenCreateFood(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "POST", Path: "/api/v1/food", Body: input}, "")
}

func (r *mutationResolver) RKitchenUpdateStatus(ctx context.Context, id string, status string) (any, error) {
	return restCall[any](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/kitchen/%v/status", id), Body: map[string]any{"status": status}}, "")
}

func (r *queryResolver) RKitchenCategories(ctx context.Context) ([]*model.RKitchenCategory, error) {
	return restCall[[]*model.RKitchenCategory](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "GET", Path: "/api/v1/category"}, "data")
}

func (r *queryResolver) RKitchenFoods(ctx context.Context, limit *int, page *int, sort *string) (*model.RKitchenFoodPage, error) {
	return restCall[*model.RKitchenFoodPage](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "GET", Path: "/api/v1/food", Query: restQuery("Limit", limit, "Page", page, "Sort", sort)}, "data")
}

func (r *queryResolver) RKitchenFoodsByIDs(ctx context.Context, ids string) ([]*model.RKitchenFood, error) {
	return restCall[[]*model.RKitchenFood](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "GET", Path: "/api/v1/food/batch", Query: restQuery("ids", ids)}, "data")
}

func (r *queryResolver) RKitchenKitchens(ctx context.Context, limit *int, page *int, sort *string) (*model.RKitchenKitchenPage, error) {
	return restCall[*model.RKitchenKitchenPage](ctx, r.Proxy, "kitchen-service", rest.Request{Method: "GET", Path: "/api/v1/kitchen", Query: restQuery("Limit", limit, "Page", page, "Sort", sort)}, "data")
}
