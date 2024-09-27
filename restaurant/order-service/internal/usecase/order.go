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

