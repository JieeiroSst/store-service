package application

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type variableService struct {
	repo port.VariableRepository
}

func NewVariableService(repo port.VariableRepository) port.VariableUsecase {
	return &variableService{repo: repo}
}

func (s *variableService) Create(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	return s.repo.Create(ctx, variable)
}

func (s *variableService) Get(ctx context.Context, key string) (*model.Variable, error) {
	return s.repo.Get(ctx, key)
}

func (s *variableService) Update(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	return s.repo.Update(ctx, variable)
}

func (s *variableService) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, key)
}

func (s *variableService) List(ctx context.Context, limit, offset int32) (*model.VariableList, error) {
	return s.repo.List(ctx, limit, offset)
}
