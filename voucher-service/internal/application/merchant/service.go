package merchant

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Service struct {
	repo  MerchantRepository
	clock shared.Clock
}

func NewService(repo MerchantRepository, clock shared.Clock) MerchantService {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) RegisterMerchant(ctx context.Context, in RegisterMerchantInput) (*merchant.Merchant, error) {
	m, err := merchant.NewMerchant(in.Name, in.ProviderType, in.Config, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) GetMerchant(ctx context.Context, id shared.MerchantID) (*merchant.Merchant, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListMerchants(ctx context.Context) ([]*merchant.Merchant, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) UpdateMerchantConfig(ctx context.Context, id shared.MerchantID, config map[string]any) (*merchant.Merchant, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	m.UpdateConfig(config, s.clock.Now())
	if err := s.repo.Save(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) DeactivateMerchant(ctx context.Context, id shared.MerchantID) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	m.Deactivate(s.clock.Now())
	return s.repo.Save(ctx, m)
}
