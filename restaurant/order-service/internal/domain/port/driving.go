package port

import (
	"context"

	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
)

type OrderUsecase interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	CancelOrder(ctx context.Context, id int, order *model.Order) error
	SuccessOrder(ctx context.Context, id int, order *model.Order) error
	FindByID(ctx context.Context, id int) (*model.Order, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type ReservationUsecase interface {
	Book(ctx context.Context, reservation *model.Reservation) error
	Cancel(ctx context.Context, id int) error
	FindByID(ctx context.Context, id int) (*model.Reservation, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}
