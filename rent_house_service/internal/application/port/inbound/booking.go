package inbound

import (
	"context"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type BookCommand struct {
	HomestayID int64
	CheckIn    time.Time
	CheckOut   time.Time
	Guests     int
	Currency   string
	Note       string
	RequestID  string
}

type BookingUseCase interface {
	Book(ctx context.Context, actor Principal, cmd BookCommand) (domain.Booking, error)
	Cancel(ctx context.Context, actor Principal, id int64) (domain.Booking, error)
	Get(ctx context.Context, actor Principal, id int64) (domain.Booking, error)
	List(ctx context.Context, actor Principal, userID int64, asHost bool, limit, offset int) ([]domain.Booking, error)
	Pay(ctx context.Context, actor Principal, id int64, cmd PayCommand) (domain.Booking, error)
	ReleaseExpired(ctx context.Context) (int, error)
}

type PayCommand struct {
	Method   domain.PaymentMethod
	Provider string
}

type WalletUseCase interface {
	Get(ctx context.Context, actor Principal) (domain.Wallet, error)
	Create(ctx context.Context, actor Principal, currency string) (domain.Wallet, error)
	Transactions(ctx context.Context, actor Principal, limit, offset int) ([]domain.WalletTxn, error)
}
