package application

import (
	"context"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
)

type rateService struct {
	rates port.RateRepository
}

func NewRateService(rates port.RateRepository) port.RateUsecase {
	return &rateService{rates: rates}
}

func (s *rateService) ListRates(ctx context.Context) ([]model.Rate, error) {
	return s.rates.List(ctx)
}
