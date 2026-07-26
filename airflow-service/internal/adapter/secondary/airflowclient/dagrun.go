package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
	"github.com/apache/airflow-client-go/airflow"
)

type dagRunRepository struct {
	client *Client
}

func NewDAGRunRepository(client *Client) port.DAGRunRepository {
	return &dagRunRepository{client: client}
}

func (r *dagRunRepository) Trigger(ctx context.Context, dagId string, req *model.TriggerDAGRunRequest) (*model.DAGRun, error) {
	body := airflow.NewDAGRun()
	if req.DagRunId != "" {
		body.SetDagRunId(req.DagRunId)
	}
	if req.LogicalDate != nil {
		body.SetLogicalDate(*req.LogicalDate)
	}
	if req.Conf != nil {
		body.SetConf(req.Conf)
	}
	if req.Note != "" {
		body.SetNote(req.Note)
	}

	resp, _, err := r.client.API.DAGRunApi.PostDagRun(r.client.Auth(ctx), dagId).DAGRun(*body).Execute()
	if err != nil {
		return nil, err
	}

	run := toDAGRunModel(resp)
	return &run, nil
}

func (r *dagRunRepository) List(ctx context.Context, dagId string, limit, offset int32) (*model.DAGRunList, error) {
	resp, _, err := r.client.API.DAGRunApi.GetDagRuns(r.client.Auth(ctx), dagId).Limit(limit).Offset(offset).Execute()
	if err != nil {
		return nil, err
	}

	runs := make([]model.DAGRun, 0, len(resp.GetDagRuns()))
	for _, run := range resp.GetDagRuns() {
		runs = append(runs, toDAGRunModel(run))
	}

	return &model.DAGRunList{DagRuns: runs, TotalEntries: resp.GetTotalEntries()}, nil
}

func (r *dagRunRepository) Get(ctx context.Context, dagId, dagRunId string) (*model.DAGRun, error) {
	resp, _, err := r.client.API.DAGRunApi.GetDagRun(r.client.Auth(ctx), dagId, dagRunId).Execute()
	if err != nil {
		return nil, err
	}

	run := toDAGRunModel(resp)
	return &run, nil
}

func (r *dagRunRepository) Delete(ctx context.Context, dagId, dagRunId string) error {
	_, err := r.client.API.DAGRunApi.DeleteDagRun(r.client.Auth(ctx), dagId, dagRunId).Execute()
	return err
}

func toDAGRunModel(d airflow.DAGRun) model.DAGRun {
	run := model.DAGRun{
		DagId:           d.GetDagId(),
		DagRunId:        d.GetDagRunId(),
		State:           string(d.GetState()),
		ExternalTrigger: d.GetExternalTrigger(),
		Conf:            d.GetConf(),
		Note:            d.GetNote(),
	}

	if logicalDate := d.GetLogicalDate(); !logicalDate.IsZero() {
		run.LogicalDate = &logicalDate
	}
	if startDate := d.GetStartDate(); !startDate.IsZero() {
		run.StartDate = &startDate
	}
	if endDate := d.GetEndDate(); !endDate.IsZero() {
		run.EndDate = &endDate
	}

	return run
}
