package order

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

// ---- Driving port (inbound use cases) ----

type CreateOrderItemInput struct {
	MerchantID shared.MerchantID
	ProductSKU string
	Quantity   int
	UnitPrice  shared.Money
}

type CreateOrderInput struct {
	BuyerType      order.BuyerType
	BuyerID        string
	Currency       string
	Items          []CreateOrderItemInput
	IdempotencyKey string
}

type CheckoutInput struct {
	OrderID        shared.OrderID
	PaymentMethod  string // "vnpay" | "momo" | "wallet"
	IdempotencyKey string
}

type CheckoutOutput struct {
	OrderID       string         `json:"order_id"`
	Status        string         `json:"status"`
	PaymentAction *PaymentIntent `json:"payment_action,omitempty"` // nil when paid via wallet synchronously
}

type OrderService interface {
	CreateOrder(ctx context.Context, in CreateOrderInput) (*order.Order, error)
	Checkout(ctx context.Context, in CheckoutInput) (*CheckoutOutput, error)
	ConfirmPayment(ctx context.Context, id shared.OrderID, paymentRef string) error
	GetOrder(ctx context.Context, id shared.OrderID) (*order.Order, error)
	ListOrders(ctx context.Context, buyerID string) ([]*order.Order, error)
	CancelOrder(ctx context.Context, id shared.OrderID, reason string) error
}

// ---- Driven ports (outbound) ----

type OrderRepository interface {
	Create(ctx context.Context, o *order.Order) error
	FindByID(ctx context.Context, id shared.OrderID) (*order.Order, error)
	FindByIDForUpdate(ctx context.Context, id shared.OrderID) (*order.Order, error)
	ListByBuyer(ctx context.Context, buyerID string) ([]*order.Order, error)
	Save(ctx context.Context, o *order.Order) error
}

type VoucherIssuanceItem struct {
	MerchantID   shared.MerchantID
	ProductSKU   string
	Denomination shared.Money
	Quantity     int
}

type VoucherIssuanceRequest struct {
	OrderID shared.OrderID
	Items   []VoucherIssuanceItem
}

type IssuedVoucherRef struct {
	VoucherID  string
	Code       string
	MerchantID string
	PIN        string
}

type VoucherIssuer interface {
	IssueVouchersForOrder(ctx context.Context, req VoucherIssuanceRequest) ([]IssuedVoucherRef, error)
}

type PaymentRequest struct {
	OrderID   shared.OrderID
	Amount    shared.Money
	Method    string
	ReturnURL string
}

type PaymentIntent struct {
	PaymentID   string `json:"payment_id"`
	RedirectURL string `json:"redirect_url"`
}

type PaymentInitiator interface {
	InitiatePayment(ctx context.Context, req PaymentRequest) (PaymentIntent, error)
}

type WalletDebiter interface {
	Debit(ctx context.Context, ownerType, ownerID string, amount shared.Money, reason string) error
}
