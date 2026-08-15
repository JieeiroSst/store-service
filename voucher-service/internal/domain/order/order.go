package order

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type BuyerType string

const (
	BuyerTypeRetail    BuyerType = "retail"
	BuyerTypeCorporate BuyerType = "corporate"
)

type Order struct {
	ID          shared.OrderID
	BuyerType   BuyerType
	BuyerID     string
	Items       []OrderItem
	Status      Status
	TotalAmount shared.Money

	Version        int
	IdempotencyKey string
	PaymentRef     string

	CreatedAt time.Time
	UpdatedAt time.Time

	PersistedVersion int

	events []shared.DomainEvent
}

func NewOrder(buyerType BuyerType, buyerID, currency, idempotencyKey string, now time.Time) (*Order, error) {
	if buyerID == "" {
		return nil, ErrInvalidOrder
	}
	return &Order{
		ID:             shared.NewOrderID(),
		BuyerType:      buyerType,
		BuyerID:        buyerID,
		Status:         StatusPending,
		TotalAmount:    shared.ZeroMoney(currency),
		Version:        1,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (o *Order) PullEvents() []shared.DomainEvent {
	events := o.events
	o.events = nil
	return events
}

func (o *Order) transitionTo(target Status, now time.Time) error {
	if !o.Status.canTransitionTo(target) {
		return ErrInvalidOrderTransition
	}
	o.Status = target
	o.Version++
	o.UpdatedAt = now
	return nil
}

func (o *Order) AddItem(item OrderItem, now time.Time) error {
	if o.Status != StatusPending {
		return ErrInvalidOrderTransition
	}
	o.Items = append(o.Items, item)
	total, err := o.TotalAmount.Add(item.LineTotal)
	if err != nil {
		return err
	}
	o.TotalAmount = total
	o.UpdatedAt = now
	return nil
}

func (o *Order) MarkAwaitingPayment(now time.Time) error {
	if len(o.Items) == 0 {
		return ErrEmptyOrder
	}
	if err := o.transitionTo(StatusAwaitingPayment, now); err != nil {
		return err
	}
	o.events = append(o.events, OrderCreatedEvent{
		BaseEvent: shared.NewBaseEvent(EventTypeOrderCreated, o.ID.String(), now),
		OrderID:   o.ID.String(),
	})
	return nil
}

func (o *Order) MarkPaid(paymentRef string, now time.Time) error {
	if err := o.transitionTo(StatusPaid, now); err != nil {
		return err
	}
	o.PaymentRef = paymentRef
	o.events = append(o.events, OrderPaidEvent{
		BaseEvent:  shared.NewBaseEvent(EventTypeOrderPaid, o.ID.String(), now),
		OrderID:    o.ID.String(),
		PaymentRef: paymentRef,
	})
	return nil
}

func (o *Order) MarkFulfilling(now time.Time) error {
	return o.transitionTo(StatusFulfilling, now)
}

func (o *Order) AttachIssuedVouchers(itemIndex int, voucherIDs []shared.VoucherID) error {
	if itemIndex < 0 || itemIndex >= len(o.Items) {
		return ErrInvalidOrder
	}
	o.Items[itemIndex].IssuedVoucherIDs = voucherIDs
	return nil
}

func (o *Order) Complete(now time.Time) error {
	if err := o.transitionTo(StatusCompleted, now); err != nil {
		return err
	}
	o.events = append(o.events, OrderFulfilledEvent{
		BaseEvent: shared.NewBaseEvent(EventTypeOrderFulfilled, o.ID.String(), now),
		OrderID:   o.ID.String(),
	})
	return nil
}

func (o *Order) Cancel(reason string, now time.Time) error {
	if err := o.transitionTo(StatusCancelled, now); err != nil {
		return err
	}
	o.events = append(o.events, OrderCancelledEvent{
		BaseEvent: shared.NewBaseEvent(EventTypeOrderCancelled, o.ID.String(), now),
		OrderID:   o.ID.String(),
		Reason:    reason,
	})
	return nil
}

func (o *Order) Fail(reason string, now time.Time) error {
	return o.transitionTo(StatusFailed, now)
}
