package application

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type poolService struct {
	repo port.PoolRepository
}

func NewPoolService(repo port.PoolRepository) port.PoolUsecase {
	return &poolService{repo: repo}
}

func (s *poolService) Create(ctx context.Context, pool *model.Pool) (*model.Pool, error) {
	return s.repo.Create(ctx, pool)
}

func (s *poolService) Get(ctx context.Context, name string) (*model.Pool, error) {
	return s.repo.Get(ctx, name)
}

func (s *poolService) Update(ctx context.Context, pool *model.Pool) (*model.Pool, error) {
	return s.repo.Update(ctx, pool)
}

func (s *poolService) Delete(ctx context.Context, name string) error {
	return s.repo.Delete(ctx, name)
}

func (s *poolService) List(ctx context.Context, limit, offset int32) (*model.PoolList, error) {
	return s.repo.List(ctx, limit, offset)
}
