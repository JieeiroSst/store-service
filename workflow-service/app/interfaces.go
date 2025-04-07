package app

type OrderRepository interface {
	GetOrder(orderID string) (*Order, error)
	UpdateOrder(order *Order) error
}

type PaymentClient interface {
	ProcessPayment(orderID string, amount float64) (string, error)
	RefundPayment(paymentID string) error
}

type FulfillmentService interface {
	FulfillOrder(orderID string) (string, error)
}

type NotificationService interface {
	SendOrderConfirmation(orderID string, trackingNumber string) error
}
