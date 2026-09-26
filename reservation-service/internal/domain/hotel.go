package domain

import (
	"fmt"
	"time"
)

const RequiredStars = 5

type HotelStatus int16

const (
	HotelActive   HotelStatus = 1 // published
	HotelInactive HotelStatus = 2 // paused by its manager
	HotelPending  HotelStatus = 3 // registered, waiting for the platform admin to verify it
	HotelRejected HotelStatus = 4 // the admin turned it down; see ReviewNote
)

func (s HotelStatus) Name() string {
	switch s {
	case HotelActive:
		return "active"
	case HotelInactive:
		return "inactive"
	case HotelPending:
		return "pending"
	case HotelRejected:
		return "rejected"
	}
	return ""
}

type Hotel struct {
	ID                int64
	Name              string
	Description       string
	Stars             int
	City              string // where guests search for it
	Address           string
	PhoneNumber       string
	Email             string
	Images            []string
	Amenities         []string
	CheckInTime       string
	CheckOutTime      string
	Currency          string
	FreeCancelHours   int
	CancellationTiers []CancellationTier
	Status            HotelStatus
	ReviewNote        string
	OwnerID           int64
	WalletID          string
	Rating            float64
	ReviewCount       int
	RoomTypes         []RoomType
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CancellationTier struct {
	HoursBefore int
	FeePercent  int
}

func (h Hotel) EffectivePolicy() []CancellationTier {
	if len(h.CancellationTiers) > 0 {
		return h.CancellationTiers
	}
	return []CancellationTier{{HoursBefore: h.FreeCancelHours, FeePercent: 0}}
}

func ValidatePolicy(tiers []CancellationTier) error {
	if len(tiers) > 5 {
		return fmt.Errorf("%w: a cancellation policy has at most 5 tiers", ErrInvalid)
	}
	for i, t := range tiers {
		if t.HoursBefore < 0 || t.HoursBefore > 24*365 || t.FeePercent < 0 || t.FeePercent > 100 {
			return fmt.Errorf("%w: tier %d is out of range (hours_before 0-8760, fee_percent 0-100)", ErrInvalid, i+1)
		}
		if i > 0 && (t.HoursBefore >= tiers[i-1].HoursBefore || t.FeePercent < tiers[i-1].FeePercent) {
			return fmt.Errorf("%w: tiers must go from earlier to later: hours_before strictly decreasing, fee_percent not decreasing", ErrInvalid)
		}
	}
	return nil
}

func FeePercentFor(tiers []CancellationTier, hoursLeft float64) int {
	for _, t := range tiers {
		if hoursLeft >= float64(t.HoursBefore) {
			return t.FeePercent
		}
	}
	return 100
}

type RoomType struct {
	ID          int64
	HotelID     int64
	Name        string
	Description string
	Capacity    int // guests per room
	Bed         string
	SizeM2      int
	View        string
	Amenities   []string
	Active      bool
}

type Room struct {
	ID         int64
	HotelID    int64
	RoomTypeID int64
	Name       string
	Floor      int
	Available  bool
}

type InventoryDay struct {
	HotelID    int64
	RoomTypeID int64
	Date       time.Time
	Total      int
	Reserved   int
	Rate       string
}

func (d InventoryDay) Left() int { return d.Total - d.Reserved }

type AddonUnit string

const (
	UnitStay  AddonUnit = "stay"  // charged once per reservation
	UnitNight AddonUnit = "night" // charged for every night
)

func (u AddonUnit) Valid() bool { return u == UnitStay || u == UnitNight }

type HotelService struct {
	ID          int64
	HotelID     int64
	Name        string
	Description string
	Price       string
	Unit        AddonUnit
	Active      bool
}

type Quote struct {
	RoomType  RoomType
	Nights    int
	Available int // rooms bookable for every night; 0 when any night is sold out or unpriced
	Total     string
}

type HotelFilter struct {
	City            string // case-insensitive match
	Query           string // matches name, city or address
	OwnerID         int64  // only this manager's hotels
	Status          HotelStatus
	IncludeInactive bool
	CheckIn         time.Time
	CheckOut        time.Time
	Rooms           int
	Guests          int
	Amenities       []string
	MinRate         string
	MaxRate         string
	Sort            string
	After           *HotelCursor
	Limit           int
	Offset          int
}

type HotelCursor struct {
	Sort   string
	ID     int64
	Rating float64
	Count  int
	Score  float64
}

type HotelPage struct {
	Items []Hotel
	Next  *HotelCursor
}

const (
	ratingPrior  = 3.5
	ratingWeight = 5.0
)

func BayesianRating(avg float64, count int) float64 {
	n := float64(count)
	return (n*avg + ratingWeight*ratingPrior) / (n + ratingWeight)
}

func TierBoost(tier int) float64 {
	switch {
	case tier >= 4:
		return 0.5
	case tier == 3:
		return 0.3
	case tier == 2:
		return 0.15
	}
	return 0
}

func RankScore(avg float64, count, tier int) float64 {
	return BayesianRating(avg, count) + TierBoost(tier)
}

func ManagerPoints(rating int) int64 {
	switch rating {
	case 5:
		return 200
	case 4:
		return 100
	}
	return 0
}

type OccupancyDay struct {
	Date     time.Time
	Rooms    int // rooms for sale that night
	Reserved int
}

type HotelReport struct {
	From, To       time.Time // nights [From, To)
	Days           []OccupancyDay
	Currency       string
	PaidCount      int
	PaidRevenue    string // total of reservations that are paid
	RefundedCount  int
	RefundedAmount string // paid back to guests
	FeesKept       string // cancellation fees the hotel kept
	StatusCounts   map[ReservationStatus]int
	NoShows        int
}
