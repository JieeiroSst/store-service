package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/JIeeiroSst/accounting-service/internal/domain/port"
)

// operatingHoursStart/End mirror the original handler's literal 8/23 gate:
// orders are accepted only between 09:00 and 22:59.
const (
	operatingHoursStart = 8
	operatingHoursEnd   = 23
)

type authCartService struct {
	repo      port.AuthCartRepository
	payments  port.PaymentRepository
	publisher port.OrderPublisher
	now       func() time.Time
}

func NewAuthCartService(repo port.AuthCartRepository, payments port.PaymentRepository, publisher port.OrderPublisher) port.AuthCartUsecase {
	return &authCartService{repo: repo, payments: payments, publisher: publisher, now: time.Now}
}

func (s *authCartService) PlaceOrder(ctx context.Context, cart model.AuthCart) error {
	hour := s.now().Hour()

	if hour > operatingHoursStart && hour < operatingHoursEnd {
		return s.acceptOrder(ctx, cart)
	}

	return s.rejectOrder(ctx, cart)
}

func (s *authCartService) acceptOrder(ctx context.Context, cart model.AuthCart) error {
	if err := s.publisher.PublishOrderSuccess(ctx, cart.Order); err != nil {
		return err
	}
	if err := s.publisher.PublishDeliveryShip(ctx, cart.Delivery); err != nil {
		return err
	}
	if err := s.repo.SaveDelivery(ctx, cart.Delivery); err != nil {
		return err
	}

	cart.Order.Status = model.OrderStatusSuccess
	if err := s.repo.SaveOrder(ctx, cart.Order); err != nil {
		return err
	}

	return s.payments.Create(ctx, model.Payment{
		OrderID: cart.Order.ID,
		Amount:  cart.Order.TotalAmount,
		Method:  methodOrDefault(cart.PaymentMethod),
		Status:  model.PaymentStatusPending,
	})
}

func methodOrDefault(method string) string {
	if method == "" {
		return model.PaymentMethodCash
	}
	return method
}

func (s *authCartService) rejectOrder(ctx context.Context, cart model.AuthCart) error {
	if err := s.publisher.PublishOrderReject(ctx, cart.Order); err != nil {
		return err
	}

	cart.Order.Status = model.OrderStatusReject
	return s.repo.SaveOrder(ctx, cart.Order)
}
