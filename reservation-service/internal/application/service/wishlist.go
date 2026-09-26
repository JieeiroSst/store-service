package service

import (
	"context"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type WishlistService struct {
	repo   outbound.WishlistRepository
	hotels outbound.HotelRepository
}

var _ inbound.WishlistUseCase = (*WishlistService)(nil)

func NewWishlistService(r outbound.WishlistRepository, h outbound.HotelRepository) *WishlistService {
	return &WishlistService{repo: r, hotels: h}
}

func (s *WishlistService) Add(ctx context.Context, a inbound.Principal, hotelID int64) error {
	h, err := s.hotels.Get(ctx, hotelID)
	if err != nil {
		return err
	}
	if h.Status != domain.HotelActive {
		return domain.ErrNotFound 
	}
	return s.repo.Add(ctx, a.UserID, hotelID)
}

func (s *WishlistService) Remove(ctx context.Context, a inbound.Principal, hotelID int64) error {
	return s.repo.Remove(ctx, a.UserID, hotelID)
}

func (s *WishlistService) List(ctx context.Context, a inbound.Principal, afterID int64, limit int) (domain.Page[domain.Hotel], error) {
	limit, _ = page(limit, 0)
	rows, err := s.repo.List(ctx, a.UserID, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Hotel]{}, err
	}
	return domain.PageOf(rows, limit, func(h domain.Hotel) int64 { return h.ID }), nil
}
