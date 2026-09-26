package inbound

import (
	"context"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type HotelInput struct {
	Name              string
	Description       string
	Stars             int // must be 5
	City              string
	Address           string
	PhoneNumber       string
	Email             string
	Images            []string
	Amenities         []string
	CheckInTime       string // "15:00"; empty keeps the default (14:00)
	CheckOutTime      string // empty keeps the default (12:00)
	Currency          string
	FreeCancelHours   *int // nil keeps the current value (48 for a new hotel)
	CancellationTiers *[]domain.CancellationTier
	WalletID          *string
	Status            domain.HotelStatus
}

type HotelUseCase interface {
	Search(ctx context.Context, f domain.HotelFilter) (domain.HotelPage, error)
	View(ctx context.Context, viewer *Principal, id int64) (domain.Hotel, error)
	Register(ctx context.Context, actor Principal, in HotelInput) (domain.Hotel, error)
	Update(ctx context.Context, actor Principal, id int64, in HotelInput) (domain.Hotel, error)
	Deactivate(ctx context.Context, actor Principal, id int64) error
	Mine(ctx context.Context, actor Principal, after *domain.HotelCursor, limit int) (domain.HotelPage, error)
	ForReview(ctx context.Context, actor Principal, status domain.HotelStatus, after *domain.HotelCursor, limit int) (domain.HotelPage, error)
	Approve(ctx context.Context, actor Principal, id int64) (domain.Hotel, error)
	Reject(ctx context.Context, actor Principal, id int64, reason string) (domain.Hotel, error)
}

type RoomTypeInput struct {
	Name        string
	Description string
	Capacity    int
	Bed         string
	SizeM2      int
	View        string
	Amenities   []string
	Active      *bool // nil = keep (true for a new one)
}

type RoomInput struct {
	RoomTypeID int64
	Name       string
	Floor      int
	Available  *bool // nil = keep (true for a new one)
}

type InventoryCommand struct {
	HotelID, RoomTypeID int64
	From, To            time.Time // nights [From, To)
	Total               *int
	Rate                *string
}

type ServiceInput struct {
	Name        string
	Description string
	Price       string
	Unit        domain.AddonUnit
	Active      *bool
}

type PromotionInput struct {
	Code        string
	Description string
	PercentOff  int    // 1-100, or 0 when AmountOff is used
	AmountOff   string // a fixed amount off the room charge, or "" when PercentOff is used
	MinNights   int    // 0 = 1
	MaxUses     int    // 0 = unlimited
	ValidFrom   *time.Time
	ValidTo     *time.Time
	Active      *bool // nil = keep (true for a new one)
}

// CatalogUseCase is what a hotel's manager maintains: room types, rooms, the nightly inventory and
// rates, and extra services. Reads are public; writes are for the hotel's manager (or an admin).
type CatalogUseCase interface {
	RoomTypes(ctx context.Context, hotelID int64) ([]domain.RoomType, error)
	CreateRoomType(ctx context.Context, actor Principal, hotelID int64, in RoomTypeInput) (domain.RoomType, error)
	UpdateRoomType(ctx context.Context, actor Principal, hotelID, id int64, in RoomTypeInput) (domain.RoomType, error)

	Rooms(ctx context.Context, actor Principal, hotelID, roomTypeID int64) ([]domain.Room, error)
	CreateRoom(ctx context.Context, actor Principal, hotelID int64, in RoomInput) (domain.Room, error)
	UpdateRoom(ctx context.Context, actor Principal, hotelID, id int64, in RoomInput) (domain.Room, error)

	SetInventory(ctx context.Context, actor Principal, c InventoryCommand) error
	Inventory(ctx context.Context, hotelID, roomTypeID int64, from, to time.Time) ([]domain.InventoryDay, error)
	// Quotes prices a stay for every room type of the hotel, and how many rooms are left.
	Quotes(ctx context.Context, hotelID int64, from, to time.Time, rooms int) ([]domain.Quote, error)

	Promotions(ctx context.Context, actor Principal, hotelID int64) ([]domain.Promotion, error)
	CreatePromotion(ctx context.Context, actor Principal, hotelID int64, in PromotionInput) (domain.Promotion, error)
	UpdatePromotion(ctx context.Context, actor Principal, hotelID, id int64, in PromotionInput) (domain.Promotion, error)

	// Report shows a manager how full the hotel was and what it earned over [from, to).
	Report(ctx context.Context, actor Principal, hotelID int64, from, to time.Time) (domain.HotelReport, error)

	Services(ctx context.Context, hotelID int64) ([]domain.HotelService, error)
	CreateService(ctx context.Context, actor Principal, hotelID int64, in ServiceInput) (domain.HotelService, error)
	UpdateService(ctx context.Context, actor Principal, hotelID, id int64, in ServiceInput) (domain.HotelService, error)
}
