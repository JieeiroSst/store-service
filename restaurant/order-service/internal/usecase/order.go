package usecase

import (
	"context"

	"github.com/JIeeiroSst/order-service/internal/dto"
	"github.com/JIeeiroSst/order-service/internal/repository"
	"github.com/JieeiroSst/logger"
)

type Orders interface {
	CreateOrder(ctx context.Context, order dto.Order) error
	CancelOrder(ctx context.Context, id int, order dto.Order) error
	SuccessOrder(ctx context.Context, id int, order dto.Order) error
	FindByID(ctx context.Context, id int) (*dto.Order, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type OrderUsecase struct {
	OrderRepository repository.Orders
}

func NewOrderUsecase(OrderRepository repository.Orders) *OrderUsecase {
	return &OrderUsecase{
		OrderRepository: OrderRepository,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, order dto.Order) error {
	model := order.CreateOrder()
	if err := u.OrderRepository.Create(ctx, model); err != nil {
		return err
	}
	return nil
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, id int, order dto.Order) error {
	model := order.CancelOrder()
	if err := u.OrderRepository.Update(ctx, id, model); err != nil {
		return err
	}
	return nil
}

func (u *OrderUsecase) SuccessOrder(ctx context.Context, id int, order dto.Order) error {
	model := order.SuccessOrder()
	if err := u.OrderRepository.Update(ctx, id, model); err != nil {
		return err
	}
	return nil
}

func (u *OrderUsecase) FindByID(ctx context.Context, id int) (*dto.Order, error) {
	order, err := u.OrderRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.BuildOrder(order), nil
}

func (u *OrderUsecase) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	orders, err := u.OrderRepository.FindAll(ctx, pagination)
	if err != nil {
		return logger.Pagination{}, nil
	}
	return orders, nil
}
