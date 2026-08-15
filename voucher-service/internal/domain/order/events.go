package order

import "github.com/JIeeiroSst/voucher-service/internal/domain/shared"

const (
	EventTypeOrderCreated    = "order.created"
	EventTypeOrderPaid       = "order.paid"
	EventTypeOrderFulfilled  = "order.fulfilled"
	EventTypeOrderCancelled  = "order.cancelled"
	EventTypeOrderFailed     = "order.failed"
)

type OrderCreatedEvent struct {
	shared.BaseEvent
	OrderID string
}

type OrderPaidEvent struct {
	shared.BaseEvent
	OrderID    string
	PaymentRef string
}

type OrderFulfilledEvent struct {
	shared.BaseEvent
	OrderID string
}

type OrderCancelledEvent struct {
	shared.BaseEvent
	OrderID string
	Reason  string
}
