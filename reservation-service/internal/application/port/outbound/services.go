package outbound

import (
	"context"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type Identity struct {
	UserID int64
	Email  string
	Roles  []string
}

type IdentityProvider interface {
	Resolve(ctx context.Context, token string) (Identity, error)
}

type TransferParams struct {
	FromWalletID string
	ToWalletID   string
	Amount       int64
	ReferenceID  string
	Description  string
}

type WalletGateway interface {
	Get(ctx context.Context, walletID string) (domain.Wallet, error)
	GetByUser(ctx context.Context, userID int64) (domain.Wallet, error)
	Create(ctx context.Context, userID int64, currency string) (domain.Wallet, error)
	Transfer(ctx context.Context, p TransferParams) (transferID string, err error)
	ReverseTransfer(ctx context.Context, transferID, reason string) error
	Transactions(ctx context.Context, walletID string, limit, offset int) ([]domain.WalletTxn, error)
}

type CreateGatewayPayment struct {
	Provider       string
	Amount         int64
	Currency       string
	PayerEmail     string
	Description    string
	IdempotencyKey string
}

type PaymentGateway interface {
	Create(ctx context.Context, in CreateGatewayPayment) (domain.GatewayPayment, error)
	Get(ctx context.Context, id int64) (domain.GatewayPayment, error)
	Refund(ctx context.Context, id int64) error
	RefundPartial(ctx context.Context, id, amount int64) error
}

type Member struct {
	Points   int64
	Tier     int // 1 bronze .. 4 platinum
	TierName string
}

type LoyaltyGateway interface {
	Earn(ctx context.Context, memberID int64, idempotencyKey string, amount int64) error
	Status(ctx context.Context, memberID int64) (Member, error)
}

type PushGateway interface {
	Push(ctx context.Context, n domain.Notification) error
}
