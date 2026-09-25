package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const maxCommentLen = 2000

type ReviewService struct {
	reviews   outbound.ReviewRepository
	homestays outbound.HomestayRepository
	loyalty   outbound.LoyaltyGateway 
	log       *slog.Logger
}

var _ inbound.ReviewUseCase = (*ReviewService)(nil)

func NewReviewService(r outbound.ReviewRepository, h outbound.HomestayRepository, l outbound.LoyaltyGateway, log *slog.Logger) *ReviewService {
	if log == nil {
		log = slog.Default()
	}
	return &ReviewService{reviews: r, homestays: h, loyalty: l, log: log}
}

func (s *ReviewService) List(ctx context.Context, homestayID int64, limit, offset int) ([]domain.Review, error) {
	if _, err := s.homestays.Get(ctx, homestayID); err != nil {
		return nil, err
	}
	limit, offset = page(limit, offset)
	return s.reviews.ListByHomestay(ctx, homestayID, limit, offset)
}

func (s *ReviewService) Create(ctx context.Context, actor inbound.Principal, c inbound.ReviewCommand) (domain.Review, error) {
	if c.Rating < 1 || c.Rating > 5 {
		return domain.Review{}, fmt.Errorf("%w: rating must be between 1 and 5", domain.ErrInvalid)
	}
	if (c.BookingID == 0) == (c.LeaseID == 0) {
		return domain.Review{}, fmt.Errorf("%w: give exactly one of booking_id or lease_id", domain.ErrInvalid)
	}
	comment := strings.TrimSpace(c.Comment)
	if len(comment) > maxCommentLen {
		return domain.Review{}, fmt.Errorf("%w: comment is longer than %d characters", domain.ErrInvalid, maxCommentLen)
	}
	h, err := s.homestays.Get(ctx, c.HomestayID)
	if err != nil {
		return domain.Review{}, err
	}
	rv, err := s.reviews.Create(ctx, domain.Review{
		HomestayID: c.HomestayID, UserID: actor.UserID, BookingID: c.BookingID, LeaseID: c.LeaseID,
		Rating: c.Rating, Comment: comment,
	})
	if err != nil {
		return domain.Review{}, err
	}

	if pts := domain.HostPoints(rv.Rating); s.loyalty != nil && pts > 0 && h.HostID != 0 {
		if err := s.loyalty.Earn(ctx, h.HostID, fmt.Sprintf("rent-house-review-%d", rv.ID), pts); err != nil {
			s.log.Warn("host points not credited", "review", rv.ID, "host", h.HostID, "err", err)
		}
	}
	return rv, nil
}

type WishlistService struct{ repo outbound.WishlistRepository }

var _ inbound.WishlistUseCase = (*WishlistService)(nil)

func NewWishlistService(r outbound.WishlistRepository) *WishlistService {
	return &WishlistService{repo: r}
}

func (s *WishlistService) Add(ctx context.Context, a inbound.Principal, id int64) error {
	return s.repo.Add(ctx, a.UserID, id)
}
func (s *WishlistService) Remove(ctx context.Context, a inbound.Principal, id int64) error {
	return s.repo.Remove(ctx, a.UserID, id)
}
func (s *WishlistService) List(ctx context.Context, a inbound.Principal, limit, offset int) ([]domain.Homestay, error) {
	limit, offset = page(limit, offset)
	return s.repo.List(ctx, a.UserID, limit, offset)
}

type LoyaltyService struct{ loyalty outbound.LoyaltyGateway }

var _ inbound.LoyaltyUseCase = (*LoyaltyService)(nil)

func NewLoyaltyService(l outbound.LoyaltyGateway) *LoyaltyService { return &LoyaltyService{loyalty: l} }

func (s *LoyaltyService) Status(ctx context.Context, a inbound.Principal) (inbound.LoyaltyStatus, error) {
	if s.loyalty == nil {
		return inbound.LoyaltyStatus{}, fmt.Errorf("%w: loyalty points are not enabled", domain.ErrInvalid)
	}
	m, err := s.loyalty.Status(ctx, a.UserID)
	if err != nil {
		return inbound.LoyaltyStatus{}, err
	}
	return inbound.LoyaltyStatus{Points: m.Points, Tier: m.Tier, TierName: m.TierName}, nil
}
