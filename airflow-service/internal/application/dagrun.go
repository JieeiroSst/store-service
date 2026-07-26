package application

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type dagRunService struct {
	repo port.DAGRunRepository
}

func NewDAGRunService(repo port.DAGRunRepository) port.DAGRunUsecase {
	return &dagRunService{repo: repo}
}

func (s *dagRunService) Trigger(ctx context.Context, dagId string, req *model.TriggerDAGRunRequest) (*model.DAGRun, error) {
	return s.repo.Trigger(ctx, dagId, req)
}

func (s *dagRunService) List(ctx context.Context, dagId string, limit, offset int32) (*model.DAGRunList, error) {
	return s.repo.List(ctx, dagId, limit, offset)
}

func (s *dagRunService) Get(ctx context.Context, dagId, dagRunId string) (*model.DAGRun, error) {
	return s.repo.Get(ctx, dagId, dagRunId)
}

func (s *dagRunService) Delete(ctx context.Context, dagId, dagRunId string) error {
	return s.repo.Delete(ctx, dagId, dagRunId)
}
