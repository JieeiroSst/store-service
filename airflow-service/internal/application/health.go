package application

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type healthService struct {
	repo port.HealthRepository
}

func NewHealthService(repo port.HealthRepository) port.HealthUsecase {
	return &healthService{repo: repo}
}

func (s *healthService) Get(ctx context.Context) (*model.HealthStatus, error) {
	return s.repo.Get(ctx)
}
