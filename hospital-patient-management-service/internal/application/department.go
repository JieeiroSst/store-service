package application

import (
	"context"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type DepartmentService struct {
	base[model.Department]
	repo port.DepartmentRepository
}

func NewDepartmentService(repo port.DepartmentRepository) port.DepartmentUsecase {
	return &DepartmentService{base: base[model.Department]{repo}, repo: repo}
}

func (s *DepartmentService) Create(ctx context.Context, d *model.Department) (*model.Department, error) {
	if err := required("name", d.Name); err != nil {
		return nil, err
	}
	return create[model.Department](ctx, s.repo, d)
}

func (s *DepartmentService) Update(ctx context.Context, d *model.Department) (*model.Department, error) {
	if err := required("name", d.Name); err != nil {
		return nil, err
	}
	return update[model.Department](ctx, s.repo, d.ID, d)
}

func (s *DepartmentService) List(ctx context.Context, page model.Page) ([]model.Department, int64, error) {
	return s.repo.List(ctx, page)
}

func (s *DepartmentService) Stats(ctx context.Context) ([]model.DepartmentStat, error) {
	return s.repo.Stats(ctx)
}
