package application

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type taskInstanceService struct {
	repo port.TaskInstanceRepository
}

func NewTaskInstanceService(repo port.TaskInstanceRepository) port.TaskInstanceUsecase {
	return &taskInstanceService{repo: repo}
}

func (s *taskInstanceService) List(ctx context.Context, dagId, dagRunId string) (*model.TaskInstanceList, error) {
	return s.repo.List(ctx, dagId, dagRunId)
}

func (s *taskInstanceService) Get(ctx context.Context, dagId, dagRunId, taskId string) (*model.TaskInstance, error) {
	return s.repo.Get(ctx, dagId, dagRunId, taskId)
}
