package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) AirflowSetDagPaused(ctx context.Context, dagID string, isPaused bool) (*model.AirflowDag, error) {
	return restCall[*model.AirflowDag](ctx, r.Proxy, "airflow-service", rest.Request{Method: "PATCH", Path: rest.P("/dags/%v", dagID), Body: map[string]any{"is_paused": isPaused}}, "data")
}

func (r *mutationResolver) AirflowTriggerDagRun(ctx context.Context, dagID string, input model.AirflowTriggerDagRunInput) (*model.AirflowDagRun, error) {
	return restCall[*model.AirflowDagRun](ctx, r.Proxy, "airflow-service", rest.Request{Method: "POST", Path: rest.P("/dags/%v/dagRuns", dagID), Body: input}, "data")
}

func (r *mutationResolver) AirflowDeleteDagRun(ctx context.Context, dagID string, dagRunID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "airflow-service", rest.Request{Method: "DELETE", Path: rest.P("/dags/%v/dagRuns/%v", dagID, dagRunID)}, "")
}

func (r *mutationResolver) AirflowCreateVariable(ctx context.Context, input model.AirflowVariableInput) (*model.AirflowVariable, error) {
	return restCall[*model.AirflowVariable](ctx, r.Proxy, "airflow-service", rest.Request{Method: "POST", Path: "/variables", Body: input}, "data")
}

func (r *mutationResolver) AirflowUpdateVariable(ctx context.Context, key string, input model.AirflowVariableInput) (*model.AirflowVariable, error) {
	return restCall[*model.AirflowVariable](ctx, r.Proxy, "airflow-service", rest.Request{Method: "PATCH", Path: rest.P("/variables/%v", key), Body: input}, "data")
}

func (r *mutationResolver) AirflowDeleteVariable(ctx context.Context, key string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "airflow-service", rest.Request{Method: "DELETE", Path: rest.P("/variables/%v", key)}, "")
}

func (r *mutationResolver) AirflowCreatePool(ctx context.Context, input model.AirflowPoolInput) (*model.AirflowPool, error) {
	return restCall[*model.AirflowPool](ctx, r.Proxy, "airflow-service", rest.Request{Method: "POST", Path: "/pools", Body: input}, "data")
}

func (r *mutationResolver) AirflowUpdatePool(ctx context.Context, name string, input model.AirflowPoolInput) (*model.AirflowPool, error) {
	return restCall[*model.AirflowPool](ctx, r.Proxy, "airflow-service", rest.Request{Method: "PATCH", Path: rest.P("/pools/%v", name), Body: input}, "data")
}

func (r *mutationResolver) AirflowDeletePool(ctx context.Context, name string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "airflow-service", rest.Request{Method: "DELETE", Path: rest.P("/pools/%v", name)}, "")
}

func (r *queryResolver) AirflowHealth(ctx context.Context) (*model.AirflowHealth, error) {
	return restCall[*model.AirflowHealth](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: "/health"}, "data")
}

func (r *queryResolver) AirflowDags(ctx context.Context, limit *int, offset *int) (*model.AirflowDagList, error) {
	return restCall[*model.AirflowDagList](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: "/dags", Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) AirflowDag(ctx context.Context, dagID string) (*model.AirflowDag, error) {
	return restCall[*model.AirflowDag](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/dags/%v", dagID)}, "data")
}

func (r *queryResolver) AirflowDagRuns(ctx context.Context, dagID string, limit *int, offset *int) (*model.AirflowDagRunList, error) {
	return restCall[*model.AirflowDagRunList](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/dags/%v/dagRuns", dagID), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) AirflowDagRun(ctx context.Context, dagID string, dagRunID string) (*model.AirflowDagRun, error) {
	return restCall[*model.AirflowDagRun](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/dags/%v/dagRuns/%v", dagID, dagRunID)}, "data")
}

func (r *queryResolver) AirflowTaskInstances(ctx context.Context, dagID string, dagRunID string, limit *int, offset *int) (*model.AirflowTaskInstanceList, error) {
	return restCall[*model.AirflowTaskInstanceList](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/dags/%v/dagRuns/%v/taskInstances", dagID, dagRunID), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) AirflowTaskInstance(ctx context.Context, dagID string, dagRunID string, taskID string) (*model.AirflowTaskInstance, error) {
	return restCall[*model.AirflowTaskInstance](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/dags/%v/dagRuns/%v/taskInstances/%v", dagID, dagRunID, taskID)}, "data")
}

func (r *queryResolver) AirflowVariables(ctx context.Context, limit *int, offset *int) (*model.AirflowVariableList, error) {
	return restCall[*model.AirflowVariableList](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: "/variables", Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) AirflowVariable(ctx context.Context, key string) (*model.AirflowVariable, error) {
	return restCall[*model.AirflowVariable](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/variables/%v", key)}, "data")
}

func (r *queryResolver) AirflowPools(ctx context.Context, limit *int, offset *int) (*model.AirflowPoolList, error) {
	return restCall[*model.AirflowPoolList](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: "/pools", Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) AirflowPool(ctx context.Context, name string) (*model.AirflowPool, error) {
	return restCall[*model.AirflowPool](ctx, r.Proxy, "airflow-service", rest.Request{Method: "GET", Path: rest.P("/pools/%v", name)}, "data")
}
