package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
	"github.com/apache/airflow-client-go/airflow"
)

type taskInstanceRepository struct {
	client *Client
}

func NewTaskInstanceRepository(client *Client) port.TaskInstanceRepository {
	return &taskInstanceRepository{client: client}
}

func (r *taskInstanceRepository) List(ctx context.Context, dagId, dagRunId string) (*model.TaskInstanceList, error) {
	resp, _, err := r.client.API.TaskInstanceApi.GetTaskInstances(r.client.Auth(ctx), dagId, dagRunId).Execute()
	if err != nil {
		return nil, err
	}

	instances := make([]model.TaskInstance, 0, len(resp.GetTaskInstances()))
	for _, ti := range resp.GetTaskInstances() {
		instances = append(instances, toTaskInstanceModel(ti))
	}

	return &model.TaskInstanceList{TaskInstances: instances, TotalEntries: resp.GetTotalEntries()}, nil
}

func (r *taskInstanceRepository) Get(ctx context.Context, dagId, dagRunId, taskId string) (*model.TaskInstance, error) {
	resp, _, err := r.client.API.TaskInstanceApi.GetTaskInstance(r.client.Auth(ctx), dagId, dagRunId, taskId).Execute()
	if err != nil {
		return nil, err
	}

	ti := toTaskInstanceModel(resp)
	return &ti, nil
}

func toTaskInstanceModel(ti airflow.TaskInstance) model.TaskInstance {
	return model.TaskInstance{
		TaskId:        ti.GetTaskId(),
		DagId:         ti.GetDagId(),
		DagRunId:      ti.GetDagRunId(),
		ExecutionDate: ti.GetExecutionDate(),
		StartDate:     ti.GetStartDate(),
		EndDate:       ti.GetEndDate(),
		Duration:      ti.GetDuration(),
		State:         string(ti.GetState()),
		TryNumber:     ti.GetTryNumber(),
		MapIndex:      ti.GetMapIndex(),
		Pool:          ti.GetPool(),
		Operator:      ti.GetOperator(),
	}
}
