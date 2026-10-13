package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ROrderCreateReservation(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "order-service", rest.Request{Method: "POST", Path: "/api/v1/reservations", Body: input}, "")
}

func (r *mutationResolver) ROrderCancelReservation(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "order-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/reservations/%v", id)}, "")
}

func (r *queryResolver) ROrderOrder(ctx context.Context, id int) (*model.ROrderOrder, error) {
	return restCall[*model.ROrderOrder](ctx, r.Proxy, "order-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/order/%v", id)}, "data")
}

func (r *queryResolver) ROrderOrders(ctx context.Context, limit *int, page *int, sort *string) (*model.ROrderOrderPage, error) {
	return restCall[*model.ROrderOrderPage](ctx, r.Proxy, "order-service", rest.Request{Method: "GET", Path: "/api/v1/order", Query: restQuery("Limit", limit, "Page", page, "Sort", sort)}, "data")
}

func (r *queryResolver) ROrderReservations(ctx context.Context, limit *int, page *int, sort *string) (*model.ROrderReservationPage, error) {
	return restCall[*model.ROrderReservationPage](ctx, r.Proxy, "order-service", rest.Request{Method: "GET", Path: "/api/v1/reservations", Query: restQuery("Limit", limit, "Page", page, "Sort", sort)}, "data")
}

func (r *queryResolver) ROrderReservation(ctx context.Context, id int) (*model.ROrderReservation, error) {
	return restCall[*model.ROrderReservation](ctx, r.Proxy, "order-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reservations/%v", id)}, "data")
}
