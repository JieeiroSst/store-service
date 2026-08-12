package project

import (
	"context"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	repo  port.ProjectRepository
	audit port.AuditService
}

func NewService(repo port.ProjectRepository, audit port.AuditService) port.ProjectService {
	return &service{repo: repo, audit: audit}
}

func (s *service) Create(ctx context.Context, name, description string, actor uuid.UUID) (*model.Project, error) {
	p := &model.Project{Name: name, Description: description, CreatedBy: actor}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, "project", p.ID, model.ActionCreate, &p.ID, nil, &actor, nil, p)
	return p, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, apperr.ErrNotFound
	}
	return p, nil
}

func (s *service) List(ctx context.Context) ([]model.Project, error) {
	return s.repo.List(ctx)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, name, description string, actor uuid.UUID) (*model.Project, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperr.ErrNotFound
	}
	before := *existing
	existing.Name = name
	existing.Description = description
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, "project", existing.ID, model.ActionUpdate, &existing.ID, nil, &actor, before, existing)
	return existing, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperr.ErrNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, "project", id, model.ActionDelete, &id, nil, &actor, existing, nil)
	return nil
}
