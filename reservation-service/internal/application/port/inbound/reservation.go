package inbound

import (
	"context"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type AddonRequest struct {
	ServiceID int64
	Quantity  int
}

type ReserveCommand struct {
	HotelID, RoomTypeID     int64
	Start, End              time.Time // check-in day, check-out day
	Rooms, Adults, Children int
	SpecialRequests         string
	Addons                  []AddonRequest
	// RequestID makes the call idempotent: retries return the original reservation.
	RequestID string
	// PromoCode is a code given out by the hotel: it takes a discount off the room charge.
	PromoCode string
}

type PayCommand struct {
	Method domain.PaymentMethod
	// Provider is the gateway provider (paypal, stripe, ...); only for MethodGateway.
	Provider string
}

type ReservationListQuery struct {
	// AsManager lists reservations at the hotels the caller manages instead of their own.
	AsManager bool
	HotelID   int64
	Status    domain.ReservationStatus
	AfterID   int64 // the cursor of the previous page
	Limit     int
}

type ReservationUseCase interface {
	Reserve(ctx context.Context, actor Principal, cmd ReserveCommand) (domain.Reservation, error)
	SoldOut(ctx context.Context, cmd ReserveCommand) bool
	Get(ctx context.Context, actor Principal, id int64) (domain.Reservation, error)
	List(ctx context.Context, actor Principal, q ReservationListQuery) (domain.Page[domain.Reservation], error)
	Pay(ctx context.Context, actor Principal, id int64, cmd PayCommand) (domain.Reservation, error)
	Cancel(ctx context.Context, actor Principal, id int64) (domain.Reservation, error)
	CancellationQuote(ctx context.Context, actor Principal, id int64) (CancellationQuote, error)
	CheckIn(ctx context.Context, actor Principal, id int64) (domain.Reservation, error)
	CheckOut(ctx context.Context, actor Principal, id int64) (domain.Reservation, error)
	NoShow(ctx context.Context, actor Principal, id int64) (domain.Reservation, error)
	History(ctx context.Context, actor Principal, id int64) ([]HistoryEntry, error)
	Reject(ctx context.Context, actor Principal, id int64, reason string) (domain.Reservation, error)
	ReleaseExpired(ctx context.Context) (int, error)
	SendReminders(ctx context.Context) (int, error)
}

type CancellationQuote struct {
	Status      domain.ReservationStatus
	HoursLeft   float64 // until check-in
	FeePercent  int
	Fee         string // kept by the hotel
	Refund      string // paid back to the guest
	Cancellable bool
	Reason      string // why not, when it is not cancellable
	Policy      []domain.CancellationTier
}

type HistoryEntry struct {
	At    time.Time
	By    string
	Event string
	From  domain.ReservationStatus
	To    domain.ReservationStatus
	Note  string
}

type WaitCommand struct {
	HotelID, RoomTypeID     int64
	Start, End              time.Time
	Rooms, Adults, Children int
}

type WaitlistUseCase interface {
	Join(ctx context.Context, actor Principal, cmd WaitCommand) (domain.WaitlistEntry, error)
	Leave(ctx context.Context, actor Principal, id int64) error
	List(ctx context.Context, actor Principal, afterID int64, limit int) (domain.Page[domain.WaitlistEntry], error)
	Promote(ctx context.Context, roomTypeID int64) (int, error)
	PromoteAll(ctx context.Context) (int, error)
}
