package domain

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

// pending -> paid | canceled | rejected;   paid -> refunded
type ReservationStatus int16

const (
	ReservationPending  ReservationStatus = 1 // rooms held, waiting for payment
	ReservationPaid     ReservationStatus = 2
	ReservationCanceled ReservationStatus = 3 // by the guest before paying, or the hold ran out
	ReservationRejected ReservationStatus = 4 // by the hotel before paying
	ReservationRefunded ReservationStatus = 5 // a paid reservation that was cancelled and paid back
)

func (s ReservationStatus) Name() string {
	switch s {
	case ReservationPending:
		return "pending"
	case ReservationPaid:
		return "paid"
	case ReservationCanceled:
		return "canceled"
	case ReservationRejected:
		return "rejected"
	case ReservationRefunded:
		return "refunded"
	}
	return ""
}

type ReservationAddon struct {
	ServiceID int64
	Name      string
	Quantity  int
	UnitPrice string
	Amount    string
}

type Reservation struct {
	ID              int64
	HotelID         int64
	RoomTypeID      int64
	GuestID         int64
	Start           time.Time
	End             time.Time // check-out day
	Rooms           int
	Adults          int
	Children        int
	Status          ReservationStatus
	Currency        string
	RoomTotal       string
	AddonsTotal     string
	PromoID         int64
	PromoCode       string
	DiscountAmount  string // taken off RoomTotal by the promo code
	TotalAmount     string // RoomTotal - DiscountAmount + AddonsTotal
	FeeAmount       string // kept by the hotel when a paid reservation was cancelled late
	RefundAmount    string // paid back to the guest when it was cancelled
	Stay            StayStatus
	CheckedInAt     *time.Time
	CheckedOutAt    *time.Time
	SpecialRequests string
	RequestID       string
	PaymentMethod   PaymentMethod
	PaymentRef      string // wallet transfer id or gateway payment id
	PaidAt          *time.Time
	ExpiresAt       *time.Time
	StatusNote      string
	Addons          []ReservationAddon
	Version         int64
	CreatedAt       time.Time
}

func (r Reservation) Nights() int { return int(r.End.Sub(r.Start).Hours() / 24) }

// AddonAmount prices an extra service: price x quantity, times the nights for per-night services.
// Amounts are decimal strings computed with exact arithmetic.
func AddonAmount(price string, unit AddonUnit, quantity, nights int) (string, error) {
	p, ok := new(big.Rat).SetString(strings.TrimSpace(price))
	if !ok || p.Sign() < 0 {
		return "", fmt.Errorf("%w: bad price %q", ErrInvalid, price)
	}
	n := int64(quantity)
	if unit == UnitNight {
		n *= int64(nights)
	}
	return new(big.Rat).Mul(p, big.NewRat(n, 1)).FloatString(4), nil
}

// SumAmounts adds decimal strings exactly.
func SumAmounts(parts ...string) (string, error) {
	sum := new(big.Rat)
	for _, s := range parts {
		v, ok := new(big.Rat).SetString(strings.TrimSpace(s))
		if !ok {
			return "", fmt.Errorf("%w: bad amount %q", ErrInvalid, s)
		}
		sum.Add(sum, v)
	}
	return sum.FloatString(4), nil
}

// StayStatus is what happened at the front desk, separate from the money.
type StayStatus int16

const (
	StayNotArrived StayStatus = 0
	StayCheckedIn  StayStatus = 1
	StayCheckedOut StayStatus = 2
	StayNoShow     StayStatus = 3
)

func (s StayStatus) Name() string {
	switch s {
	case StayNotArrived:
		return "not_arrived"
	case StayCheckedIn:
		return "checked_in"
	case StayCheckedOut:
		return "checked_out"
	case StayNoShow:
		return "no_show"
	}
	return ""
}

func FeeAmount(total string, feePercent int, currency string) (string, error) {
	minor, err := MinorUnits(total, currency)
	if err != nil {
		return "", err
	}
	fee := (minor*int64(feePercent) + 50) / 100 // round half up
	return FromMinorUnits(fee, currency), nil
}

func CanRefundWithoutFee(now, start time.Time, freeCancelHours int) bool {
	return now.Before(start.Add(-time.Duration(freeCancelHours) * time.Hour))
}

type Review struct {
	ID            int64
	HotelID       int64
	UserID        int64
	ReservationID int64
	Rating        int
	Comment       string
	CreatedAt     time.Time
}

// ---- history, promotions, waiting list, notifications

type ReservationEvent struct {
	ID            int64
	ReservationID int64
	At            time.Time
	Actor         int64 // 0 = the system
	Event         string
	FromStatus    ReservationStatus
	ToStatus      ReservationStatus
	Note          string
}

type Promotion struct {
	ID          int64
	HotelID     int64
	Code        string
	Description string
	PercentOff  int    // 1-100, or 0 when AmountOff is used
	AmountOff   string // a fixed amount, or "" when PercentOff is used
	MinNights   int
	MaxUses     int // 0 = unlimited
	UsedCount   int
	ValidFrom   *time.Time
	ValidTo     *time.Time
	Active      bool
}

func (p Promotion) DiscountFor(roomTotal string) (string, error) {
	total, ok := new(big.Rat).SetString(strings.TrimSpace(roomTotal))
	if !ok {
		return "", fmt.Errorf("%w: bad amount %q", ErrInvalid, roomTotal)
	}
	var d *big.Rat
	if p.PercentOff > 0 {
		d = new(big.Rat).Mul(total, big.NewRat(int64(p.PercentOff), 100))
	} else {
		v, ok := new(big.Rat).SetString(strings.TrimSpace(p.AmountOff))
		if !ok {
			return "", fmt.Errorf("%w: bad amount %q", ErrInvalid, p.AmountOff)
		}
		d = v
	}
	if d.Cmp(total) > 0 {
		d = total
	}
	return d.FloatString(4), nil
}

func (p Promotion) UsableFor(start time.Time, nights int) string {
	switch {
	case !p.Active:
		return "the promo code is not active"
	case p.MaxUses > 0 && p.UsedCount >= p.MaxUses:
		return "the promo code has been used up"
	case p.ValidFrom != nil && start.Before(*p.ValidFrom):
		return "the promo code is not valid for those dates yet"
	case p.ValidTo != nil && start.After(*p.ValidTo):
		return "the promo code has expired for those dates"
	case nights < p.MinNights:
		return fmt.Sprintf("the promo code needs a stay of at least %d nights", p.MinNights)
	}
	return ""
}

type WaitlistStatus int16

const (
	WaitlistWaiting   WaitlistStatus = 1
	WaitlistOffered   WaitlistStatus = 2
	WaitlistCancelled WaitlistStatus = 3
)

type WaitlistEntry struct {
	ID            int64
	HotelID       int64
	RoomTypeID    int64
	GuestID       int64
	Start         time.Time
	End           time.Time
	Rooms         int
	Adults        int
	Children      int
	Status        WaitlistStatus
	ReservationID int64
	CreatedAt     time.Time
	OfferedAt     *time.Time
	OfferStatus   ReservationStatus
}

type Notification struct {
	ID            int64
	UserID        int64
	Kind          string
	Title         string
	Body          string
	ReservationID int64
	HotelID       int64
	CreatedAt     time.Time
	ReadAt        *time.Time
}

func MulAmount(a string, n int) (string, error) {
	v, ok := new(big.Rat).SetString(strings.TrimSpace(a))
	if !ok {
		return "", fmt.Errorf("%w: bad amount %q", ErrInvalid, a)
	}
	return new(big.Rat).Mul(v, big.NewRat(int64(n), 1)).FloatString(4), nil
}

func SubAmounts(a, b string) (string, error) {
	x, ok1 := new(big.Rat).SetString(strings.TrimSpace(a))
	y, ok2 := new(big.Rat).SetString(strings.TrimSpace(b))
	if !ok1 || !ok2 {
		return "", fmt.Errorf("%w: bad amount %q or %q", ErrInvalid, a, b)
	}
	return new(big.Rat).Sub(x, y).FloatString(4), nil
}
