package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) SardOnboardCustomer(ctx context.Context, input map[string]any) (*model.SardCustomer, error) {
	return restCall[*model.SardCustomer](ctx, r.Proxy, "sard-customer-info-service", rest.Request{Method: "POST", Path: "/api/v1/customers", Body: input}, "")
}

func (r *mutationResolver) SardSyncCustomer(ctx context.Context, id string) (*model.SardCustomer, error) {
	return restCall[*model.SardCustomer](ctx, r.Proxy, "sard-customer-info-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/customers/%v/sync", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) SardSubmitKyc(ctx context.Context, id string, input map[string]any) (*model.SardCustomer, error) {
	return restCall[*model.SardCustomer](ctx, r.Proxy, "sard-customer-info-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/customers/%v/kyc", id), Body: input}, "")
}

func (r *mutationResolver) SardRefreshKyc(ctx context.Context, id string) (*model.SardCustomer, error) {
	return restCall[*model.SardCustomer](ctx, r.Proxy, "sard-customer-info-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/customers/%v/kyc/refresh", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) SardCustomer(ctx context.Context, id string) (*model.SardCustomer, error) {
	return restCall[*model.SardCustomer](ctx, r.Proxy, "sard-customer-info-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/customers/%v", id)}, "")
}

func (r *queryResolver) SardCustomerByUser(ctx context.Context, userID int) (*model.SardCustomer, error) {
	return restCall[*model.SardCustomer](ctx, r.Proxy, "sard-customer-info-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/customer", userID)}, "")
}
