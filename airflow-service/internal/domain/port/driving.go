package port

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
)

type DAGUsecase interface {
	List(ctx context.Context, limit, offset int32) (*model.DAGList, error)
	Get(ctx context.Context, dagId string) (*model.DAG, error)
	SetPaused(ctx context.Context, dagId string, isPaused bool) (*model.DAG, error)
}

type DAGRunUsecase interface {
	Trigger(ctx context.Context, dagId string, req *model.TriggerDAGRunRequest) (*model.DAGRun, error)
	List(ctx context.Context, dagId string, limit, offset int32) (*model.DAGRunList, error)
	Get(ctx context.Context, dagId, dagRunId string) (*model.DAGRun, error)
	Delete(ctx context.Context, dagId, dagRunId string) error
}

type TaskInstanceUsecase interface {
	List(ctx context.Context, dagId, dagRunId string) (*model.TaskInstanceList, error)
	Get(ctx context.Context, dagId, dagRunId, taskId string) (*model.TaskInstance, error)
}

type VariableUsecase interface {
	Create(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	Get(ctx context.Context, key string) (*model.Variable, error)
	Update(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, limit, offset int32) (*model.VariableList, error)
}

type PoolUsecase interface {
	Create(ctx context.Context, pool *model.Pool) (*model.Pool, error)
	Get(ctx context.Context, name string) (*model.Pool, error)
	Update(ctx context.Context, pool *model.Pool) (*model.Pool, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context, limit, offset int32) (*model.PoolList, error)
}

type HealthUsecase interface {
	Get(ctx context.Context) (*model.HealthStatus, error)
}
