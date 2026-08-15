package corporate

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/corporate"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Service struct {
	repo  CorporateRepository
	clock shared.Clock
}

func NewService(repo CorporateRepository, clock shared.Clock) CorporateService {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) RegisterCorporate(ctx context.Context, in RegisterCorporateInput) (*corporate.Corporate, error) {
	c, err := corporate.NewCorporate(in.Name, in.TaxCode, in.ContactEmail, in.BudgetLimit, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetCorporate(ctx context.Context, id shared.CorporateID) (*corporate.Corporate, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) SetBudget(ctx context.Context, id shared.CorporateID, limit shared.Money) (*corporate.Corporate, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.SetBudget(limit, s.clock.Now())
	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) CheckBudget(ctx context.Context, id shared.CorporateID, proposedSpend shared.Money) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !c.IsActive() {
		return corporate.ErrCorporateInactive
	}
	spent, err := s.repo.SpentThisPeriod(ctx, id)
	if err != nil {
		return err
	}
	return c.CheckBudget(spent, proposedSpend)
}
