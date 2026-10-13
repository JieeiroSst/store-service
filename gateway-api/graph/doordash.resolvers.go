package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) DoordashPlaceOrder(ctx context.Context, input model.DoordashOrderInput) (*model.DoordashOrder, error) {
	return restCall[*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: "/api/v1/orders", Body: withField(input, "customer_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) DoordashCancelMyOrder(ctx context.Context, id int, reason *string) (*model.DoordashOrder, error) {
	return restOwned[*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/cancel", id), Body: map[string]any{"reason": reason}}, "", "customer_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}

func (r *mutationResolver) DoordashUpdateOrderStatus(ctx context.Context, id int, status string) (*model.DoordashOrder, error) {
	return restCall[*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/orders/%v/status", id), Body: map[string]any{"status": status}}, "")
}

func (r *mutationResolver) DoordashCancelOrder(ctx context.Context, id int, reason *string) (*model.DoordashOrder, error) {
	return restCall[*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/cancel", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) DoordashAssignDriver(ctx context.Context, id int, driverID string) (*model.DoordashAssignment, error) {
	return restCall[*model.DoordashAssignment](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/assign-driver", id), Body: map[string]any{"driver_id": driverID}}, "")
}

func (r *mutationResolver) DoordashAddTrackingEvent(ctx context.Context, id int, input model.DoordashTrackingInput) (*model.DoordashTracking, error) {
	return restCall[*model.DoordashTracking](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/tracking", id), Body: input}, "")
}

func (r *mutationResolver) DoordashAcceptAssignment(ctx context.Context, id int) (*model.DoordashAssignment, error) {
	return restCall[*model.DoordashAssignment](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/driver-assignments/%v/accept", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) DoordashRejectAssignment(ctx context.Context, id int, reason *string) (*model.DoordashAssignment, error) {
	return restCall[*model.DoordashAssignment](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/driver-assignments/%v/reject", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) DoordashCompleteAssignment(ctx context.Context, id int) (*model.DoordashAssignment, error) {
	return restCall[*model.DoordashAssignment](ctx, r.Proxy, "doordash-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/driver-assignments/%v/complete", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) MyDoordashOrders(ctx context.Context) ([]*model.DoordashOrder, error) {
	return restCall[[]*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: "/api/v1/orders", Query: restQuery("customer_id", userIDOf(ctx))}, "")
}

func (r *queryResolver) DoordashMyOrder(ctx context.Context, id int) (*model.DoordashOrder, error) {
	return restOwned[*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "", "customer_id", nil, "")
}

func (r *queryResolver) DoordashMyOrderTracking(ctx context.Context, id int) ([]*model.DoordashTracking, error) {
	return restOwned[[]*model.DoordashTracking](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v/tracking", id)}, "", "customer_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}

func (r *queryResolver) DoordashOrders(ctx context.Context, customerID *string, restaurantID *string) ([]*model.DoordashOrder, error) {
	return restCall[[]*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: "/api/v1/orders", Query: restQuery("customer_id", customerID, "restaurant_id", restaurantID)}, "")
}

func (r *queryResolver) DoordashOrder(ctx context.Context, id int) (*model.DoordashOrder, error) {
	return restCall[*model.DoordashOrder](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}

func (r *queryResolver) DoordashOrderTracking(ctx context.Context, id int) ([]*model.DoordashTracking, error) {
	return restCall[[]*model.DoordashTracking](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v/tracking", id)}, "")
}

func (r *queryResolver) DoordashDriverAssignments(ctx context.Context, driverID string) ([]*model.DoordashAssignment, error) {
	return restCall[[]*model.DoordashAssignment](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: "/api/v1/driver-assignments", Query: restQuery("driver_id", driverID)}, "")
}

func (r *queryResolver) MyDriverAssignments(ctx context.Context) ([]*model.DoordashAssignment, error) {
	return restCall[[]*model.DoordashAssignment](ctx, r.Proxy, "doordash-service", rest.Request{Method: "GET", Path: "/api/v1/driver-assignments", Query: restQuery("driver_id", userIDOf(ctx))}, "")
}
