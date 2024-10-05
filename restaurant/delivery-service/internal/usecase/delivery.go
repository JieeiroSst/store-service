package usecase

import (
	"context"

	"github.com/JIeeiroSst/delivery-service/internal/dto"
	"github.com/JIeeiroSst/delivery-service/internal/repository"
	"github.com/JieeiroSst/logger"
)

type Deliveries interface {
	Create(ctx context.Context, delivery dto.Delivery) error
	UpdateStatus(ctx context.Context, shipID int, status string) error
	FindByActive(ctx context.Context) (*dto.Delivery, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	Update(ctx context.Context, shipID int, delivery dto.Delivery) error
}

type DeliveryUsecase struct {
	DeliveryRepository repository.Deliveries
}

func NewDeliveryRepository(DeliveryRepository repository.Deliveries) *DeliveryUsecase {
	return &DeliveryUsecase{
		DeliveryRepository: DeliveryRepository,
	}
}

func (u *DeliveryUsecase) Create(ctx context.Context, delivery dto.Delivery) error {
	model := delivery.Create()
	if err := u.DeliveryRepository.Create(ctx, model); err != nil {
		return err
	}
	return nil
}

func (u *DeliveryUsecase) UpdateStatus(ctx context.Context, shipID int, status string) error {
	if err := u.DeliveryRepository.UpdateStatus(ctx, shipID, shipID); err != nil {
		return err
	}
	return nil
}

func (u *DeliveryUsecase) FindByActive(ctx context.Context) (*dto.Delivery, error) {
	delivery, err := u.DeliveryRepository.FindByActive(ctx)
	if err != nil {
		return nil, err
	}
	if delivery == nil {
		return nil, err
	}

	return dto.BuildDelivery(delivery), nil
}

func (u *DeliveryUsecase) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	deliveries, err := u.DeliveryRepository.FindAll(ctx, pagination)
	if err != nil {
		return logger.Pagination{}, nil
	}
	return deliveries, nil
}

func (u *DeliveryUsecase) Update(ctx context.Context, shipID int, delivery dto.Delivery) error {
	model := delivery.Update()
	if err := u.DeliveryRepository.Update(ctx, shipID, model); err != nil {
		return err
	}
	return nil
}
