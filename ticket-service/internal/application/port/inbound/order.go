package inbound

import (
	"context"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type ReserveItem struct {
	TicketTypeID int64
	Quantity     int
	SeatIDs      []int64
}

type ReserveCommand struct {
	EventID    int64
	Items      []ReserveItem
	PromoCode  string
	BuyerName  string
	BuyerEmail string
	BuyerPhone string
	RequestID  string
}

type PayCommand struct {
	Method   domain.PaymentMethod
	Provider string
}

type OrderListQuery struct {
	EventID int64
	Status  domain.OrderStatus
	AfterID int64
	Limit   int
}

type InviteCommand struct {
	TicketTypeID int64
	Quantity     int
	SeatIDs      []int64
	UserID       int64
	Name         string
	Email        string
	Note         string
}

type OrderProgress struct {
	// Status is pending, paid, expired, cancelled, refunded, or gone (no such order).
	Status         string
	ExpiresAt      time.Time
	RetryAfter     time.Duration
	EventStartsAt  time.Time
	EventCancelled bool
}

type OrderUseCase interface {
	SoldOut(ctx context.Context, c ReserveCommand) bool
	Reserve(ctx context.Context, actor Principal, c ReserveCommand) (domain.Order, error)
	Get(ctx context.Context, actor Principal, id int64) (domain.Order, error)
	Invoice(ctx context.Context, actor Principal, id int64) (domain.Invoice, error)
	List(ctx context.Context, actor Principal, q OrderListQuery) (domain.Page[domain.Order], error)
	Pay(ctx context.Context, actor Principal, id int64, c PayCommand) (domain.Order, error)
	Cancel(ctx context.Context, actor Principal, id int64) (domain.Order, error)
	Invite(ctx context.Context, actor Principal, eventID int64, c InviteCommand) (domain.Order, error)
	Advance(ctx context.Context, id int64) (OrderProgress, error)
	Remind(ctx context.Context, id int64) (bool, error)
	ReleaseExpired(ctx context.Context) (int, error)
	SettleCancelledEvents(ctx context.Context) (int, error)
	SendReminders(ctx context.Context) (int, error)
}
