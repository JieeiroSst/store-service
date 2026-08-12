package environment

import (
	"context"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	repo port.EnvironmentRepository
}

func NewService(repo port.EnvironmentRepository) port.EnvironmentService {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, name string, envType model.EnvironmentType, sortOrder int) (*model.Environment, error) {
	e := &model.Environment{Name: name, Type: envType, SortOrder: sortOrder, Enabled: true}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*model.Environment, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, apperr.ErrNotFound
	}
	return e, nil
}

func (s *service) List(ctx context.Context) ([]model.Environment, error) {
	return s.repo.List(ctx)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, name string, envType model.EnvironmentType, enabled bool) (*model.Environment, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperr.ErrNotFound
	}
	existing.Name = name
	existing.Type = envType
	existing.Enabled = enabled
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperr.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}
