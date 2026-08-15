package reporting

import (
	"context"
	"time"
)

type Service struct {
	repo ReportingRepository
}

func NewService(repo ReportingRepository) ReportingService {
	return &Service{repo: repo}
}

func (s *Service) RedemptionRateByMerchant(ctx context.Context, from, to time.Time) ([]RedemptionReport, error) {
	return s.repo.RedemptionRateByMerchant(ctx, from, to)
}

func (s *Service) CorporateSpend(ctx context.Context, corporateID string, from, to time.Time) (*CorporateSpendReport, error) {
	return s.repo.CorporateSpend(ctx, corporateID, from, to)
}
