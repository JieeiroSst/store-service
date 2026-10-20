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

func (r *mutationResolver) CallCenterSeedDefaults(ctx context.Context) (*model.CallCenterSeedResult, error) {
	return restCall[*model.CallCenterSeedResult](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "POST", Path: "/api/scenarios/seed", Body: map[string]any{}}, "")
}

func (r *mutationResolver) CallCenterCreateKnowledgeEntry(ctx context.Context, input model.CallCenterKnowledgeInput) (*model.CallCenterKnowledgeEntry, error) {
	return restCall[*model.CallCenterKnowledgeEntry](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "POST", Path: "/api/knowledge", Body: input}, "")
}

func (r *mutationResolver) CallCenterUpdateKnowledgeEntry(ctx context.Context, entryID int, input model.CallCenterKnowledgeInput) (*model.CallCenterKnowledgeEntry, error) {
	return restCall[*model.CallCenterKnowledgeEntry](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "PUT", Path: rest.P("/api/knowledge/%v", entryID), Body: input}, "")
}

func (r *mutationResolver) CallCenterDeleteKnowledgeEntry(ctx context.Context, entryID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "DELETE", Path: rest.P("/api/knowledge/%v", entryID)}, "")
}

func (r *mutationResolver) CallCenterUpdateCustomer(ctx context.Context, phoneNumber string, input model.CallCenterCustomerInput) (*model.CallCenterCustomer, error) {
	return restCall[*model.CallCenterCustomer](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "PUT", Path: rest.P("/api/customers/%v", phoneNumber), Body: input}, "")
}

func (r *mutationResolver) CallCenterSimulate(ctx context.Context, input model.CallCenterSimulateInput) (*model.CallCenterSimulation, error) {
	return restCall[*model.CallCenterSimulation](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "POST", Path: "/api/simulate", Body: input}, "")
}

func (r *queryResolver) CallCenterCalls(ctx context.Context, skip *int, limit *int, status *string, fromNumber *string, outcome *string) ([]*model.CallCenterCall, error) {
	return restCall[[]*model.CallCenterCall](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/calls", Query: restQuery("skip", skip, "limit", limit, "status", status, "from_number", fromNumber, "outcome", outcome)}, "")
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

func (r *queryResolver) CallCenterKnowledgeEntries(ctx context.Context, skip *int, limit *int, category *string, scenario *string) ([]*model.CallCenterKnowledgeEntry, error) {
	return restCall[[]*model.CallCenterKnowledgeEntry](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/knowledge", Query: restQuery("skip", skip, "limit", limit, "category", category, "scenario", scenario)}, "")
}

func (r *queryResolver) CallCenterKnowledgeEntry(ctx context.Context, entryID int) (*model.CallCenterKnowledgeEntry, error) {
	return restCall[*model.CallCenterKnowledgeEntry](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: rest.P("/api/knowledge/%v", entryID)}, "")
}

func (r *queryResolver) CallCenterSearchKnowledge(ctx context.Context, q string, scenario *string, limit *int) ([]*model.CallCenterKnowledgeEntry, error) {
	return restCall[[]*model.CallCenterKnowledgeEntry](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/knowledge/search", Query: restQuery("q", q, "scenario", scenario, "limit", limit)}, "")
}

func (r *queryResolver) CallCenterCustomer(ctx context.Context, phoneNumber string) (*model.CallCenterCustomer, error) {
	return restCall[*model.CallCenterCustomer](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: rest.P("/api/customers/%v", phoneNumber)}, "")
}

func (r *queryResolver) CallCenterStatistics(ctx context.Context, days *int) (*model.CallCenterStatistics, error) {
	return restCall[*model.CallCenterStatistics](ctx, r.Proxy, "call-center-ai", rest.Request{Method: "GET", Path: "/api/analytics/statistics", Query: restQuery("days", days)}, "")
}
