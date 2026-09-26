package inbound

import (
	"context"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type ReviewCommand struct {
	HotelID       int64
	ReservationID int64
	Rating        int
	Comment       string
}

type ReviewUseCase interface {
	List(ctx context.Context, hotelID, afterID int64, limit int) (domain.Page[domain.Review], error)
	Create(ctx context.Context, actor Principal, cmd ReviewCommand) (domain.Review, error)
}

type LoyaltyStatus struct {
	Points   int64
	Tier     int
	TierName string
}

type LoyaltyUseCase interface {
	Status(ctx context.Context, actor Principal) (LoyaltyStatus, error)
}

type WalletUseCase interface {
	Get(ctx context.Context, actor Principal) (domain.Wallet, error)
	Create(ctx context.Context, actor Principal, currency string) (domain.Wallet, error)
	Transactions(ctx context.Context, actor Principal, limit, offset int) ([]domain.WalletTxn, error)
}

type NotificationUseCase interface {
	List(ctx context.Context, actor Principal, unreadOnly bool, afterID int64, limit int) (domain.Page[domain.Notification], error)
	UnreadCount(ctx context.Context, actor Principal) (int, error)
	MarkRead(ctx context.Context, actor Principal, id int64) error
	MarkAllRead(ctx context.Context, actor Principal) (int, error)
	PushPending(ctx context.Context) (int, error)
}

type WishlistUseCase interface {
	Add(ctx context.Context, actor Principal, hotelID int64) error
	Remove(ctx context.Context, actor Principal, hotelID int64) error
	List(ctx context.Context, actor Principal, afterID int64, limit int) (domain.Page[domain.Hotel], error)
}
