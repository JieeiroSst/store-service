package application

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/domain/model"
	"github.com/JIeeiroSst/consumer-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

type consumerService struct {
	repo      port.ConsumerRepository
	publisher port.OrderPublisher
}

func NewConsumerService(repo port.ConsumerRepository, publisher port.OrderPublisher) port.ConsumerUsecase {
	return &consumerService{repo: repo, publisher: publisher}
}

func (s *consumerService) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	return s.repo.Find(ctx, pagination)
}

func (s *consumerService) Create(ctx context.Context, consumer *model.Consumer) error {
	return s.repo.Create(ctx, *consumer)
}

// PlaceOrder fans the order out to the kitchen, order-book and accounting
// services over NATS without persisting a Consumer row itself, mirroring
// the original handler's behavior plus the accounting.authorize edge shown
// in restaurant.drawio.png (consumer-service -> accounting-service).
func (s *consumerService) PlaceOrder(ctx context.Context, order model.PlaceOrder) error {
	if err := s.publisher.PublishKitchenCreate(ctx, order); err != nil {
		return err
	}

	if err := s.publisher.PublishOrderCreated(ctx, buildOrder(order)); err != nil {
		return err
	}

	return s.publisher.PublishCartAuthorization(ctx, buildCartAuthorization(order))
}

// buildOrder carries line items rather than a total — order-service prices
// them itself from kitchen-service's catalog.
func buildOrder(o model.PlaceOrder) model.Order {
	items := make([]model.OrderItem, 0, len(o.Menu))
	for _, v := range o.Menu {
		items = append(items, model.OrderItem{FoodID: v.ID, Quantity: quantityOrDefault(v.Quantity)})
	}

	return model.Order{
		ID:        o.OrderID,
		TableName: o.Name,
		KitchenID: o.KitchenID,
		Items:     items,
	}
}

func quantityOrDefault(quantity int) int {
	if quantity <= 0 {
		return 1
	}
	return quantity
}

func buildCartAuthorization(o model.PlaceOrder) model.CartAuthorization {
	var total float64
	for _, v := range o.Menu {
		total += v.Money * float64(quantityOrDefault(v.Quantity))
	}

	return model.CartAuthorization{
		Order: model.AuthOrder{
			ID:          o.OrderID,
			TableName:   o.Name,
			KitchenID:   o.KitchenID,
			TotalAmount: total,
		},
		Delivery: model.AuthDelivery{
			Name:      o.DeliveryName,
			Address:   o.DeliveryAddress,
			KitchenID: o.KitchenID,
		},
		PaymentMethod: o.PaymentMethod,
	}
}
