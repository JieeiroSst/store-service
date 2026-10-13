package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CallCenterDeleteCall(ctx context.Context, callID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "DELETE", Path: rest.P("/api/calls/%v", callID)}, "")
}

func (r *mutationResolver) CallCenterCreateScenario(ctx context.Context, input model.CallCenterScenarioInput) (*model.CallCenterScenario, error) {
	return restCall[*model.CallCenterScenario](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "POST", Path: "/api/scenarios", Body: input}, "")
}

func (r *mutationResolver) CallCenterUpdateScenario(ctx context.Context, scenarioID int, input model.CallCenterScenarioInput) (*model.CallCenterScenario, error) {
	return restCall[*model.CallCenterScenario](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "PUT", Path: rest.P("/api/scenarios/%v", scenarioID), Body: input}, "")
}

func (r *mutationResolver) CallCenterDeleteScenario(ctx context.Context, scenarioID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "DELETE", Path: rest.P("/api/scenarios/%v", scenarioID)}, "")
}

func (r *queryResolver) CallCenterCalls(ctx context.Context, skip *int, limit *int, status *string, fromNumber *string) ([]*model.CallCenterCall, error) {
	return restCall[[]*model.CallCenterCall](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/calls", Query: restQuery("skip", skip, "limit", limit, "status", status, "from_number", fromNumber)}, "")
}

func (r *queryResolver) CallCenterCall(ctx context.Context, callID int) (*model.CallCenterCallDetail, error) {
	return restCall[*model.CallCenterCallDetail](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: rest.P("/api/calls/%v", callID)}, "")
}

func (r *queryResolver) CallCenterScenarios(ctx context.Context, skip *int, limit *int) ([]*model.CallCenterScenario, error) {
	return restCall[[]*model.CallCenterScenario](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/scenarios", Query: restQuery("skip", skip, "limit", limit)}, "")
}

func (r *queryResolver) CallCenterScenario(ctx context.Context, scenarioID int) (*model.CallCenterScenario, error) {
	return restCall[*model.CallCenterScenario](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: rest.P("/api/scenarios/%v", scenarioID)}, "")
}

func (r *queryResolver) CallCenterCustomer(ctx context.Context, phoneNumber string) (*model.CallCenterCustomer, error) {
	return restCall[*model.CallCenterCustomer](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: rest.P("/api/customers/%v", phoneNumber)}, "")
}

func (r *queryResolver) CallCenterStatistics(ctx context.Context, days *int) (*model.CallCenterStatistics, error) {
	return restCall[*model.CallCenterStatistics](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/analytics/statistics", Query: restQuery("days", days)}, "")
}
