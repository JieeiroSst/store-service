package service

import (
	"context"
	"strings"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type CatalogService struct {
	events outbound.EventRepository
	seats  *ttlCache[int64, domain.SeatMap]
	maps   *ttlCache[int64, domain.EventSeatMap]
}

var _ inbound.CatalogUseCase = (*CatalogService)(nil)

func NewCatalogService(events outbound.EventRepository, seatTTL time.Duration) *CatalogService {
	return &CatalogService{events: events, seats: newTTLCache[int64, domain.SeatMap](seatTTL, 1_000), maps: newTTLCache[int64, domain.EventSeatMap](seatTTL, 1_000)}
}

func (s *CatalogService) Search(ctx context.Context, f domain.EventFilter) ([]domain.Event, error) {
	f.Query, f.City = trim(f.Query), trim(f.City)
	f.Category = strings.ToLower(trim(f.Category))
	if f.Category != "" && !domain.ValidCategory(f.Category) {
		return nil, invalid("category must be one of %s", strings.Join(domain.Categories, ", "))
	}
	switch f.Sort {
	case "", "upcoming", "popular", "newest":
	default:
		return nil, invalid("sort must be upcoming, popular or newest")
	}
	if f.From != nil && f.To != nil && !f.To.After(*f.From) {
		return nil, invalid("'to' must be after 'from'")
	}
	f.Limit = page(f.Limit)
	f.Offset = max(f.Offset, 0)
	return s.events.Search(ctx, f)
}

func (s *CatalogService) EventSeatMap(ctx context.Context, eventID int64) (domain.EventSeatMap, error) {
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return domain.EventSeatMap{}, err
	}
	if e.Status != domain.EventPublished {
		return domain.EventSeatMap{}, domain.ErrNotFound
	}
	return s.maps.Get(eventID, func() (domain.EventSeatMap, error) { return s.events.EventSeatMap(ctx, eventID) })
}

func (s *CatalogService) Series(ctx context.Context, eventID int64) ([]domain.Event, error) {
	return s.events.Series(ctx, eventID)
}

func (s *CatalogService) Seats(ctx context.Context, eventID, typeID int64) (domain.SeatMap, error) {
	t, err := s.events.TicketType(ctx, typeID)
	if err != nil {
		return domain.SeatMap{}, err
	}
	if t.EventID != eventID || !t.Seated || !t.Active {
		return domain.SeatMap{}, domain.ErrNotFound
	}
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return domain.SeatMap{}, err
	}
	if e.Status != domain.EventPublished {
		return domain.SeatMap{}, domain.ErrNotFound
	}
	return s.seats.Get(typeID, func() (domain.SeatMap, error) {
		seats, err := s.events.Seats(ctx, typeID)
		if err != nil {
			return domain.SeatMap{}, err
		}
		layout, err := s.events.SeatLayout(ctx, typeID)
		return domain.SeatMap{Layout: layout, Seats: seats}, err
	})
}
