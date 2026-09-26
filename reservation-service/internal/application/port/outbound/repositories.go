package outbound

import (
	"context"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type SetInventoryParams struct {
	HotelID, RoomTypeID int64
	From, To            time.Time // nights [From, To)
	Total               *int
	Rate                *string
}

type HotelRepository interface {
	Create(ctx context.Context, h domain.Hotel) (domain.Hotel, error) // OwnerID and Status come with h
	Update(ctx context.Context, h domain.Hotel) (domain.Hotel, error)
	Get(ctx context.Context, id int64) (domain.Hotel, error) // with its room types
	List(ctx context.Context, f domain.HotelFilter) ([]domain.Hotel, error)
	SetStatus(ctx context.Context, id int64, s domain.HotelStatus) error
	SetReview(ctx context.Context, id int64, s domain.HotelStatus, note string) error

	CreateRoomType(ctx context.Context, rt domain.RoomType) (domain.RoomType, error) // ErrConflict on a duplicate name
	UpdateRoomType(ctx context.Context, rt domain.RoomType) (domain.RoomType, error)
	GetRoomType(ctx context.Context, id int64) (domain.RoomType, error)

	CreateRoom(ctx context.Context, r domain.Room) (domain.Room, error)
	UpdateRoom(ctx context.Context, r domain.Room) (domain.Room, error)
	GetRoom(ctx context.Context, id int64) (domain.Room, error)
	Rooms(ctx context.Context, hotelID, roomTypeID int64) ([]domain.Room, error) // roomTypeID 0 = all

	SetInventory(ctx context.Context, p SetInventoryParams) error
	Inventory(ctx context.Context, hotelID, roomTypeID int64, from, to time.Time) ([]domain.InventoryDay, error)
	Quotes(ctx context.Context, hotelID int64, from, to time.Time, rooms int) ([]domain.Quote, error)

	SavePromotion(ctx context.Context, p domain.Promotion) (domain.Promotion, error) // create when ID is 0; ErrConflict on a duplicate code
	Promotions(ctx context.Context, hotelID int64) ([]domain.Promotion, error)
	PromotionByCode(ctx context.Context, hotelID int64, code string) (domain.Promotion, error) // case-insensitive; ErrNotFound

	Report(ctx context.Context, hotelID int64, from, to time.Time) (domain.HotelReport, error)

	SaveService(ctx context.Context, s domain.HotelService) (domain.HotelService, error) // create when ID is 0
	Services(ctx context.Context, hotelID int64, activeOnly bool) ([]domain.HotelService, error)
}

type ReserveParams struct {
	GuestID, HotelID, RoomTypeID int64
	Start, End                   time.Time
	Rooms, Adults, Children      int
	Currency                     string
	SpecialRequests              string
	RequestID                    string
	ExpiresAt                    time.Time
	Addons                       []domain.ReservationAddon
	AddonsTotal                  string
	MaxPending                   int
	Promo                        *domain.Promotion
	Actor                        int64
}

type TransitionParams struct {
	ID          int64
	From        []domain.ReservationStatus
	To          domain.ReservationStatus
	Note        string
	Actor       int64
	Fee, Refund string
}

type ReservationFilter struct {
	AfterID int64 // keyset paging: only reservations with a smaller id (newest first)
	GuestID int64 // 0 = any guest
	OwnerID int64 // only reservations at this manager's hotels; 0 = any
	HotelID int64
	Status  domain.ReservationStatus // 0 = any
	Limit   int
	Offset  int
}

type ReservationRepository interface {
	Reserve(ctx context.Context, p ReserveParams) (domain.Reservation, error)
	Get(ctx context.Context, id int64) (domain.Reservation, error) // with add-ons
	RoomsLeft(ctx context.Context, hotelID, roomTypeID int64, start, end time.Time) (int, error)
	RequestIDExists(ctx context.Context, requestID string) (bool, error)
	FindByRequestID(ctx context.Context, guestID int64, requestID string) (domain.Reservation, bool, error)
	List(ctx context.Context, f ReservationFilter) ([]domain.Reservation, error)
	SetPaymentAttempt(ctx context.Context, id int64, method domain.PaymentMethod, ref string) (domain.Reservation, error)
	MarkPaid(ctx context.Context, id int64, method domain.PaymentMethod, ref string) (domain.Reservation, error)
	Transition(ctx context.Context, p TransitionParams) (domain.Reservation, error)
	ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Reservation, error)
	SetStay(ctx context.Context, id int64, from, to domain.StayStatus, actor int64) (domain.Reservation, error)
	Events(ctx context.Context, id int64) ([]domain.ReservationEvent, error)
	DueForReminder(ctx context.Context, from, to time.Time, limit int) ([]domain.Reservation, error)
	MarkReminded(ctx context.Context, id int64) (bool, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, r domain.Review) (domain.Review, error)
	ListByHotel(ctx context.Context, hotelID, afterID int64, limit int) ([]domain.Review, error)
}

type NotificationRepository interface {
	Add(ctx context.Context, n domain.Notification) error
	List(ctx context.Context, userID int64, unreadOnly bool, afterID int64, limit int) ([]domain.Notification, error)
	UnreadCount(ctx context.Context, userID int64) (int, error)
	MarkRead(ctx context.Context, userID, id int64) error
	MarkAllRead(ctx context.Context, userID int64) (int, error)
	ClaimUnpushed(ctx context.Context, limit int, lease time.Duration) ([]domain.Notification, error)
	MarkPushed(ctx context.Context, id int64) error
}

type WishlistRepository interface {
	Add(ctx context.Context, userID, hotelID int64) error
	Remove(ctx context.Context, userID, hotelID int64) error
	List(ctx context.Context, userID, afterID int64, limit int) ([]domain.Hotel, error)
}

type WaitlistRepository interface {
	Join(ctx context.Context, e domain.WaitlistEntry, maxWaiting int) (domain.WaitlistEntry, error)
	Leave(ctx context.Context, guestID, id int64) error
	List(ctx context.Context, guestID, afterID int64, limit int) ([]domain.WaitlistEntry, error)
	Waiting(ctx context.Context, roomTypeID int64, today time.Time, limit int) ([]domain.WaitlistEntry, error)
	RoomTypesWaiting(ctx context.Context, today time.Time, limit int) ([]int64, error)
	MarkOffered(ctx context.Context, id, reservationID int64) (OfferResult, error)
	ExpireStale(ctx context.Context, today time.Time) (int, error)
}

type OfferResult int

const (
	OfferMade    OfferResult = iota // recorded now
	OfferAlready                    // another replica recorded this very offer a moment ago
	OfferGone                       // the guest left the list (or it was closed) in the meantime
)
