package outbound

import (
	"context"
	"io"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type Identity struct {
	UserID int64
	Email  string
	Roles  []string
}

type IdentityProvider interface {
	Resolve(ctx context.Context, token string) (Identity, error)
}

type UserDirectory interface {
	Lookup(ctx context.Context, userID int64) (Identity, error)
}

type TransferParams struct {
	FromWalletID string
	ToWalletID   string
	Amount       int64
	ReferenceID  string
	Description  string
}

type WalletGateway interface {
	GetByUser(ctx context.Context, userID int64) (domain.Wallet, error)
	Transfer(ctx context.Context, p TransferParams) (transferID string, err error)
	ReverseTransfer(ctx context.Context, transferID, reason string) error
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
}

type PushGateway interface {
	Push(ctx context.Context, n domain.Notification) error
}

type EmailGateway interface {
	Email(ctx context.Context, n domain.Notification) error
}

type DocumentRenderer interface {
	Invoice(inv domain.Invoice, loc *time.Location) ([]byte, error)
	Tickets(doc domain.TicketDocument) ([]byte, error)
}

type DocumentStore interface {
	Put(ctx context.Context, userID int64, fileName string, pdf []byte) (fileID string, err error)
	Open(ctx context.Context, fileID string) (body io.ReadCloser, size int64, err error)
	Delete(ctx context.Context, fileID string) error
}

type OrderLifecycle interface {
	Track(ctx context.Context, orderID int64, expiresAt time.Time) error
	Nudge(ctx context.Context, orderID int64) error
}
