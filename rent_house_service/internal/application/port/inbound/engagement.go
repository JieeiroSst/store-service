package inbound

import (
	"context"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type ReviewCommand struct {
	HomestayID int64
	BookingID  int64
	LeaseID    int64
	Rating     int
	Comment    string
}

type ReviewUseCase interface {
	List(ctx context.Context, homestayID int64, limit, offset int) ([]domain.Review, error)
	Create(ctx context.Context, actor Principal, cmd ReviewCommand) (domain.Review, error)
}

type WishlistUseCase interface {
	Add(ctx context.Context, actor Principal, homestayID int64) error
	Remove(ctx context.Context, actor Principal, homestayID int64) error
	List(ctx context.Context, actor Principal, limit, offset int) ([]domain.Homestay, error)
}

type LoyaltyStatus struct {
	Points   int64
	Tier     int
	TierName string
}

type LoyaltyUseCase interface {
	Status(ctx context.Context, actor Principal) (LoyaltyStatus, error)
}
