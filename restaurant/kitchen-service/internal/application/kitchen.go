package application

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/common"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

// validPrepStatus mirrors the order a kitchen ticket moves through:
// received -> preparing -> ready -> served.
var validPrepStatus = map[string]bool{
	model.PrepStatusReceived:  true,
	model.PrepStatusPreparing: true,
	model.PrepStatusReady:     true,
	model.PrepStatusServed:    true,
}

type kitchenService struct {
	repo port.KitchenRepository
}

func NewKitchenService(repo port.KitchenRepository) port.KitchenUsecase {
	return &kitchenService{repo: repo}
}

func (s *kitchenService) Create(ctx context.Context, kitchen *model.Kitchen) error {
	kitchen.ID = logger.GearedIntID()
	kitchen.Status = model.PrepStatusReceived

	return s.repo.Create(ctx, *kitchen)
}

func (s *kitchenService) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	return s.repo.Find(ctx, pagination)
}

func (s *kitchenService) UpdateStatus(ctx context.Context, id int, status string) error {
	if !validPrepStatus[status] {
		return common.ErrInvalidPrepStatus
	}

	return s.repo.UpdateStatus(ctx, id, status)
}
