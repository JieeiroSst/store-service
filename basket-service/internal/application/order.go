package application

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type orderService struct {
	repo port.OrderRepository
}

func NewOrderService(repo port.OrderRepository) port.OrderUsecase {
	return &orderService{repo: repo}
}

func (s *orderService) GetOrder(ctx context.Context, id int) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context) ([]model.Order, error) {
	return s.repo.List(ctx)
}
