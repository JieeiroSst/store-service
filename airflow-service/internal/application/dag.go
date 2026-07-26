package application

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type dagService struct {
	repo port.DAGRepository
}

func NewDAGService(repo port.DAGRepository) port.DAGUsecase {
	return &dagService{repo: repo}
}

func (s *dagService) List(ctx context.Context, limit, offset int32) (*model.DAGList, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *dagService) Get(ctx context.Context, dagId string) (*model.DAG, error) {
	return s.repo.Get(ctx, dagId)
}

func (s *dagService) SetPaused(ctx context.Context, dagId string, isPaused bool) (*model.DAG, error) {
	return s.repo.SetPaused(ctx, dagId, isPaused)
}
