package domain

import "time"

type PaymentMethod string

const (
	MethodWallet  PaymentMethod = "wallet"  // internal e-wallet (payment-wallet-service)
	MethodGateway PaymentMethod = "gateway" // external provider (payment_service)
	MethodFree    PaymentMethod = "free"    // nothing to pay
)

type Wallet struct {
	ID       string
	UserID   int64
	Balance  int64
	Currency string
	Status   string
}

type GatewayStatus string

const (
	GatewayPending    GatewayStatus = "pending"
	GatewayAuthorized GatewayStatus = "authorized"
	GatewayCaptured   GatewayStatus = "captured"
	GatewayFailed     GatewayStatus = "failed"
	GatewayRefunded   GatewayStatus = "refunded"
	GatewayPartially  GatewayStatus = "partially_refunded"
)

type GatewayPayment struct {
	ID       int64
	Provider string
	Status   GatewayStatus
	Amount   int64
	Currency string
}

type Email struct {
	To       string
	Template string
	Data     map[string]string
}

type Notification struct {
	ID        int64
	UserID    int64
	Kind      string
	Title     string
	Body      string
	OrderID   int64
	EventID   int64
	CreatedAt time.Time
	ReadAt    *time.Time
	Email     *Email
}
