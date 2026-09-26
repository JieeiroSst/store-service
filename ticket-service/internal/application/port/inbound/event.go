package inbound

import (
	"context"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type EventInput struct {
	Title             string
	Description       string
	Category          string
	City              string
	Venue             string
	Address           string
	BannerURL         string
	StartsAt          time.Time
	EndsAt            time.Time
	Currency          string
	WalletID          string
	RefundCutoffHours int
	Transferable      bool
	ResaleCapPercent  int
}

type TicketTypeInput struct {
	Name         string
	Description  string
	Price        int64
	Total        int
	MinPerOrder  int
	MaxPerOrder  int
	MaxPerUser   int
	SaleStartsAt *time.Time
	SaleEndsAt   *time.Time
	Seated       bool
	Active       bool
	SortOrder    int
	SessionID    int64
}

type SeatLayout struct {
	Section          string
	Rows             int
	SeatsPerRow      int
	FirstRow         string
	OffsetX, OffsetY float64
}

type PromotionInput struct {
	Code       string
	Kind       domain.PromoKind
	Value      int64
	MaxUses    int
	MinTickets int
	ValidFrom  *time.Time
	ValidTo    *time.Time
	Active     bool
}

type DuplicateInput struct {
	Title          string
	StartsAt       time.Time
	EndsAt         time.Time
	CopyPromotions bool
}

type SessionInput struct {
	StartsAt time.Time
	EndsAt   time.Time
	Label    string
	CopyFrom int64
}

type EventUseCase interface {
	Create(ctx context.Context, actor Principal, in EventInput) (domain.Event, error)
	Update(ctx context.Context, actor Principal, id int64, in EventInput) (domain.Event, error)
	Duplicate(ctx context.Context, actor Principal, id int64, in DuplicateInput) (domain.Event, error)
	CreateSession(ctx context.Context, actor Principal, eventID int64, in SessionInput) (domain.Session, error)
	UpdateSession(ctx context.Context, actor Principal, eventID, sessionID int64, in SessionInput) (domain.Session, error)
	DeleteSession(ctx context.Context, actor Principal, eventID, sessionID int64) error
	CancelSession(ctx context.Context, actor Principal, eventID, sessionID int64, reason string) (domain.Session, error)
	Submit(ctx context.Context, actor Principal, id int64) (domain.Event, error)
	Approve(ctx context.Context, actor Principal, id int64) (domain.Event, error)
	Reject(ctx context.Context, actor Principal, id int64, note string) (domain.Event, error)
	Cancel(ctx context.Context, actor Principal, id int64, reason string) (domain.Event, error)
	SetFeatured(ctx context.Context, actor Principal, id int64, featured bool) (domain.Event, error)
	View(ctx context.Context, viewer *Principal, id int64) (domain.Event, error)
	ListMine(ctx context.Context, actor Principal, afterID int64, limit int) (domain.Page[domain.Event], error)
	ReviewQueue(ctx context.Context, actor Principal, afterID int64, limit int) (domain.Page[domain.Event], error)

	CreateTicketType(ctx context.Context, actor Principal, eventID int64, in TicketTypeInput) (domain.TicketType, error)
	UpdateTicketType(ctx context.Context, actor Principal, eventID, typeID int64, in TicketTypeInput) (domain.TicketType, error)
	GenerateSeats(ctx context.Context, actor Principal, eventID, typeID int64, l SeatLayout) (int, error)
	ApplyVenueMap(ctx context.Context, actor Principal, eventID int64, in ApplyVenueInput) (ApplyVenueResult, error)
	SetSeatMap(ctx context.Context, actor Principal, eventID, typeID int64, l domain.SeatMapLayout, seats []domain.SeatPosition) (int, error)

	ListPromotions(ctx context.Context, actor Principal, eventID int64) ([]domain.Promotion, error)
	CreatePromotion(ctx context.Context, actor Principal, eventID int64, in PromotionInput) (domain.Promotion, error)
	UpdatePromotion(ctx context.Context, actor Principal, eventID, promoID int64, in PromotionInput) (domain.Promotion, error)

	Report(ctx context.Context, actor Principal, eventID int64) (domain.EventReport, error)
	Attendees(ctx context.Context, actor Principal, eventID, afterID int64, limit int) (domain.Page[domain.Attendee], error)
}

type CatalogUseCase interface {
	Search(ctx context.Context, f domain.EventFilter) ([]domain.Event, error)
	EventSeatMap(ctx context.Context, eventID int64) (domain.EventSeatMap, error)
	Series(ctx context.Context, eventID int64) ([]domain.Event, error)
	Seats(ctx context.Context, eventID, typeID int64) (domain.SeatMap, error)
}

type GateUseCase interface {
	CheckIn(ctx context.Context, actor Principal, eventID int64, code string) (domain.Ticket, error)
	BatchCheckIn(ctx context.Context, actor Principal, eventID int64, items []ScanItem) ([]ScanResult, error)
}
