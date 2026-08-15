package order

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (*order.Order, error) {
	if len(in.Items) == 0 {
		return nil, order.ErrEmptyOrder
	}

	now := s.clock.Now()
	o, err := order.NewOrder(in.BuyerType, in.BuyerID, in.Currency, in.IdempotencyKey, now)
	if err != nil {
		return nil, err
	}

	for _, item := range in.Items {
		orderItem, err := order.NewOrderItem(item.MerchantID, item.ProductSKU, item.Quantity, item.UnitPrice)
		if err != nil {
			return nil, err
		}
		if err := o.AddItem(orderItem, now); err != nil {
			return nil, err
		}
	}

	if err := o.MarkAwaitingPayment(now); err != nil {
		return nil, err
	}

	err = s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, o); err != nil {
			return err
		}
		return s.enqueueEvents(ctx, o.PullEvents())
	})
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (s *Service) GetOrder(ctx context.Context, id shared.OrderID) (*order.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListOrders(ctx context.Context, buyerID string) ([]*order.Order, error) {
	return s.repo.ListByBuyer(ctx, buyerID)
}
