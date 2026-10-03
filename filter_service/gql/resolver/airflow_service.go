package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AirflowQuery() generated.AirflowQueryResolver { return &airflowQueryResolver{r} }

type airflowQueryResolver struct{ *Resolver }

func (r *airflowQueryResolver) Health(ctx context.Context, obj *model.AirflowQuery) (*model.AirflowHealthStatus, error) {
	return r.Clients.AirflowService.Health(ctx)
}

func (r *airflowQueryResolver) Dags(ctx context.Context, obj *model.AirflowQuery, limit *int, offset *int) (*model.AirflowDAGList, error) {
	return r.Clients.AirflowService.Dags(ctx, limit, offset)
}

func (r *airflowQueryResolver) Dag(ctx context.Context, obj *model.AirflowQuery, dagID string) (*model.AirflowDag, error) {
	return r.Clients.AirflowService.Dag(ctx, dagID)
}

func (r *airflowQueryResolver) DagRuns(ctx context.Context, obj *model.AirflowQuery, dagID string, limit *int, offset *int) (*model.AirflowDAGRunList, error) {
	return r.Clients.AirflowService.DagRuns(ctx, dagID, limit, offset)
}

func (r *airflowQueryResolver) DagRun(ctx context.Context, obj *model.AirflowQuery, dagID string, dagRunID string) (*model.AirflowDAGRun, error) {
	return r.Clients.AirflowService.DagRun(ctx, dagID, dagRunID)
}

func (r *airflowQueryResolver) TaskInstances(ctx context.Context, obj *model.AirflowQuery, dagID string, dagRunID string) (*model.AirflowTaskInstanceList, error) {
	return r.Clients.AirflowService.TaskInstances(ctx, dagID, dagRunID)
}

func (r *airflowQueryResolver) TaskInstance(ctx context.Context, obj *model.AirflowQuery, dagID string, dagRunID string, taskID string) (*model.AirflowTaskInstance, error) {
	return r.Clients.AirflowService.TaskInstance(ctx, dagID, dagRunID, taskID)
}

func (r *airflowQueryResolver) Variables(ctx context.Context, obj *model.AirflowQuery, limit *int, offset *int) (*model.AirflowVariableList, error) {
	return r.Clients.AirflowService.Variables(ctx, limit, offset)
}

func (r *airflowQueryResolver) Variable(ctx context.Context, obj *model.AirflowQuery, key string) (*model.AirflowVariable, error) {
	return r.Clients.AirflowService.Variable(ctx, key)
}

func (r *airflowQueryResolver) Pools(ctx context.Context, obj *model.AirflowQuery, limit *int, offset *int) (*model.AirflowPoolList, error) {
	return r.Clients.AirflowService.Pools(ctx, limit, offset)
}

func (r *airflowQueryResolver) Pool(ctx context.Context, obj *model.AirflowQuery, name string) (*model.AirflowPool, error) {
	return r.Clients.AirflowService.Pool(ctx, name)
}
