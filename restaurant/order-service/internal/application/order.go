package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/order-service/common"
	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JIeeiroSst/order-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

type orderService struct {
	repo   port.OrderRepository
	prices port.FoodPricer
}

func NewOrderService(repo port.OrderRepository, prices port.FoodPricer) port.OrderUsecase {
	return &orderService{repo: repo, prices: prices}
}

// CreateOrder prices every line item from kitchen-service's catalog rather
// than trusting whatever unit price/total a caller sends, then sums them
// into Order.TotalAmount.
func (s *orderService) CreateOrder(ctx context.Context, order *model.Order) error {
	order.ID = logger.GearedIntID()
	order.Status = model.OrderStatusPending
	order.CreatedDate = time.Now()
	order.UpdatedTime = time.Now()

	if len(order.Items) > 0 {
		foodIDs := make([]int, len(order.Items))
		for i, item := range order.Items {
			foodIDs[i] = item.FoodID
		}

		priced, err := s.prices.GetPrices(ctx, foodIDs)
		if err != nil {
			return err
		}

		var total float64
		for i, item := range order.Items {
			food, ok := priced[item.FoodID]
			if !ok {
				return common.ErrUnknownFoodInOrder
			}
			order.Items[i].FoodName = food.Name
			order.Items[i].UnitPrice = food.Price
			total += food.Price * float64(item.Quantity)
		}
		order.TotalAmount = total
	}

	return s.repo.Create(ctx, order)
}

func (s *orderService) CancelOrder(ctx context.Context, id int, order *model.Order) error {
	order.Status = model.OrderStatusCancel
	order.UpdatedTime = time.Now()

	return s.repo.Update(ctx, id, order)
}

func (s *orderService) SuccessOrder(ctx context.Context, id int, order *model.Order) error {
	order.Status = model.OrderStatusSuccess
	order.UpdatedTime = time.Now()

	return s.repo.Update(ctx, id, order)
}

func (s *orderService) FindByID(ctx context.Context, id int) (*model.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *orderService) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	return s.repo.FindAll(ctx, pagination)
}
