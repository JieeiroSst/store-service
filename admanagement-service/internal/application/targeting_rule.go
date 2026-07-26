package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adTargetingRuleService struct {
	repo port.AdTargetingRuleRepository
}

func NewAdTargetingRuleService(repo port.AdTargetingRuleRepository) port.AdTargetingRuleUsecase {
	return &adTargetingRuleService{repo: repo}
}

func (s *adTargetingRuleService) Create(ctx context.Context, rule *model.AdTargetingRule) error {
	return s.repo.Create(ctx, rule)
}

func (s *adTargetingRuleService) Get(ctx context.Context, id uint) (*model.AdTargetingRule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adTargetingRuleService) Update(ctx context.Context, rule *model.AdTargetingRule) error {
	return s.repo.Update(ctx, rule)
}

func (s *adTargetingRuleService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *adTargetingRuleService) List(ctx context.Context) ([]model.AdTargetingRule, error) {
	return s.repo.List(ctx)
}
