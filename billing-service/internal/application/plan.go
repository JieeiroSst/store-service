package application

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
)

type planService struct {
	repo port.PlanRepository
}

func NewPlanService(repo port.PlanRepository) port.PlanUsecase {
	return &planService{repo: repo}
}

func (s *planService) Create(ctx context.Context, plan *model.Plan) error {
	return s.repo.Create(ctx, plan)
}

func (s *planService) Get(ctx context.Context, id int) (*model.Plan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *planService) Update(ctx context.Context, plan *model.Plan) error {
	return s.repo.Update(ctx, plan)
}

func (s *planService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *planService) List(ctx context.Context) ([]model.Plan, error) {
	return s.repo.List(ctx)
}
