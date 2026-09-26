package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const maxCommentLen = 2000

type ReviewService struct {
	reviews outbound.ReviewRepository
	hotels  outbound.HotelRepository
	loyalty outbound.LoyaltyGateway // nil: managers earn no points
	log     *slog.Logger
}

var _ inbound.ReviewUseCase = (*ReviewService)(nil)

func NewReviewService(r outbound.ReviewRepository, h outbound.HotelRepository, l outbound.LoyaltyGateway, log *slog.Logger) *ReviewService {
	return &ReviewService{reviews: r, hotels: h, loyalty: l, log: defaultLog(log)}
}

func (s *ReviewService) List(ctx context.Context, hotelID, afterID int64, limit int) (domain.Page[domain.Review], error) {
	h, err := s.hotels.Get(ctx, hotelID)
	if err != nil {
		return domain.Page[domain.Review]{}, err
	}
	if h.Status != domain.HotelActive {
		return domain.Page[domain.Review]{}, domain.ErrNotFound
	}
	limit, _ = page(limit, 0)
	rows, err := s.reviews.ListByHotel(ctx, hotelID, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Review]{}, err
	}
	return domain.PageOf(rows, limit, func(r domain.Review) int64 { return r.ID }), nil
}

func (s *ReviewService) Create(ctx context.Context, actor inbound.Principal, c inbound.ReviewCommand) (domain.Review, error) {
	if c.Rating < 1 || c.Rating > 5 {
		return domain.Review{}, fmt.Errorf("%w: rating must be between 1 and 5", domain.ErrInvalid)
	}
	if c.ReservationID <= 0 {
		return domain.Review{}, fmt.Errorf("%w: reservation_id is required", domain.ErrInvalid)
	}
	comment := trim(c.Comment)
	if len(comment) > maxCommentLen {
		return domain.Review{}, fmt.Errorf("%w: comment is longer than %d characters", domain.ErrInvalid, maxCommentLen)
	}
	h, err := s.hotels.Get(ctx, c.HotelID)
	if err != nil {
		return domain.Review{}, err
	}
	rv, err := s.reviews.Create(ctx, domain.Review{HotelID: c.HotelID, UserID: actor.UserID, ReservationID: c.ReservationID, Rating: c.Rating, Comment: comment})
	if err != nil {
		return domain.Review{}, err
	}

	if pts := domain.ManagerPoints(rv.Rating); s.loyalty != nil && pts > 0 && h.OwnerID != 0 {
		if err := s.loyalty.Earn(ctx, h.OwnerID, fmt.Sprintf("reservation-review-%d", rv.ID), pts); err != nil {
			s.log.Warn("manager points not credited", "review", rv.ID, "manager", h.OwnerID, "err", err)
		}
	}
	return rv, nil
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
