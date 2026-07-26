package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
	"github.com/apache/airflow-client-go/airflow"
)

type dagRepository struct {
	client *Client
}

func NewDAGRepository(client *Client) port.DAGRepository {
	return &dagRepository{client: client}
}

func (r *dagRepository) List(ctx context.Context, limit, offset int32) (*model.DAGList, error) {
	resp, _, err := r.client.API.DAGApi.GetDags(r.client.Auth(ctx)).Limit(limit).Offset(offset).Execute()
	if err != nil {
		return nil, err
	}

	dags := make([]model.DAG, 0, len(resp.GetDags()))
	for _, d := range resp.GetDags() {
		dags = append(dags, toDAGModel(d))
	}

	return &model.DAGList{Dags: dags, TotalEntries: resp.GetTotalEntries()}, nil
}

func (r *dagRepository) Get(ctx context.Context, dagId string) (*model.DAG, error) {
	resp, _, err := r.client.API.DAGApi.GetDag(r.client.Auth(ctx), dagId).Execute()
	if err != nil {
		return nil, err
	}

	dag := toDAGModel(resp)
	return &dag, nil
}

func (r *dagRepository) SetPaused(ctx context.Context, dagId string, isPaused bool) (*model.DAG, error) {
	patch := airflow.NewDAG()
	patch.SetIsPaused(isPaused)

	resp, _, err := r.client.API.DAGApi.PatchDag(r.client.Auth(ctx), dagId).
		DAG(*patch).
		UpdateMask([]string{"is_paused"}).
		Execute()
	if err != nil {
		return nil, err
	}

	dag := toDAGModel(resp)
	return &dag, nil
}

func toDAGModel(d airflow.DAG) model.DAG {
	return model.DAG{
		DagId:       d.GetDagId(),
		IsPaused:    d.GetIsPaused(),
		IsActive:    d.GetIsActive(),
		IsSubdag:    d.GetIsSubdag(),
		Owners:      d.GetOwners(),
		Description: d.GetDescription(),
		Fileloc:     d.GetFileloc(),
	}
}
