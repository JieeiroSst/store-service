package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CallCenterQuery() generated.CallCenterQueryResolver {
	return &callCenterQueryResolver{r}
}

type callCenterQueryResolver struct{ *Resolver }

func (r *callCenterQueryResolver) Calls(ctx context.Context, obj *model.CallCenterQuery, skip *int, limit *int, status *string, fromNumber *string) ([]*model.CallCenterCall, error) {
	return r.Clients.CallCenterAi.Calls(ctx, skip, limit, status, fromNumber)
}

func (r *callCenterQueryResolver) Call(ctx context.Context, obj *model.CallCenterQuery, callID int) (*model.CallCenterCallHistory, error) {
	return r.Clients.CallCenterAi.Call(ctx, callID)
}

func (r *callCenterQueryResolver) Scenarios(ctx context.Context, obj *model.CallCenterQuery, skip *int, limit *int) ([]*model.CallCenterScenario, error) {
	return r.Clients.CallCenterAi.Scenarios(ctx, skip, limit)
}

func (r *callCenterQueryResolver) Scenario(ctx context.Context, obj *model.CallCenterQuery, scenarioID int) (*model.CallCenterScenario, error) {
	return r.Clients.CallCenterAi.Scenario(ctx, scenarioID)
}

func (r *callCenterQueryResolver) Customer(ctx context.Context, obj *model.CallCenterQuery, phoneNumber string) (*model.CallCenterCustomer, error) {
	return r.Clients.CallCenterAi.Customer(ctx, phoneNumber)
}

func (r *callCenterQueryResolver) Statistics(ctx context.Context, obj *model.CallCenterQuery, days *int) (*model.CallCenterCallStatistics, error) {
	return r.Clients.CallCenterAi.Statistics(ctx, days)
}
