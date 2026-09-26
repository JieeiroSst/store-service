package domain

import (
	"strconv"
	"strings"
	"time"
)

type EventStatus int16

const (
	EventDraft     EventStatus = 1
	EventPending   EventStatus = 2 // submitted, waiting for the platform admin
	EventPublished EventStatus = 3
	EventRejected  EventStatus = 4
	EventCancelled EventStatus = 5
)

func (s EventStatus) Name() string {
	switch s {
	case EventDraft:
		return "draft"
	case EventPending:
		return "pending_review"
	case EventPublished:
		return "published"
	case EventRejected:
		return "rejected"
	case EventCancelled:
		return "cancelled"
	}
	return "unknown"
}

func ParseEventStatus(s string) EventStatus {
	for _, st := range []EventStatus{EventDraft, EventPending, EventPublished, EventRejected, EventCancelled} {
		if st.Name() == s {
			return st
		}
	}
	return 0
}

var Categories = []string{"music", "arts", "sports", "workshop", "other"}

func ValidCategory(c string) bool {
	for _, x := range Categories {
		if x == c {
			return true
		}
	}
	return false
}

type Event struct {
	ID                int64
	OrganizerID       int64
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
	Status            EventStatus
	ReviewNote        string
	Featured          bool
	Transferable      bool
	ResaleCapPercent  int
	SeriesID          int64
	VenueID           int64
	WalletID          string
	RefundCutoffHours int
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Sessions          []Session
	TicketTypes       []TicketType
	NextSession       *time.Time
	MinPrice          *int64
	SoldOut           bool
}

func (e Event) OnSale(now time.Time) bool {
	return e.Status == EventPublished && now.Before(e.EndsAt)
}

func (e Event) RefundableAt(start, now time.Time) bool {
	return e.RefundCutoffHours > 0 && now.Add(time.Duration(e.RefundCutoffHours)*time.Hour).Before(start)
}

func (e Event) Refundable(now time.Time) bool {
	return e.RefundCutoffHours > 0 && now.Add(time.Duration(e.RefundCutoffHours)*time.Hour).Before(e.StartsAt)
}

type EventFilter struct {
	Query    string
	City     string
	Category string
	From, To *time.Time
	MinPrice *int64
	MaxPrice *int64
	Featured bool
	Sort     string // "upcoming" (default), "popular", "newest"
	Limit    int
	Offset   int
}

type TicketType struct {
	ID           int64
	EventID      int64
	Name         string
	Description  string
	Price        int64
	Total        int
	Available    int
	Sold         int
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

func (t TicketType) Held() int { return t.Total - t.Available - t.Sold }

func (t TicketType) OnSale(now time.Time) bool {
	if !t.Active {
		return false
	}
	if t.SaleStartsAt != nil && now.Before(*t.SaleStartsAt) {
		return false
	}
	return t.SaleEndsAt == nil || now.Before(*t.SaleEndsAt)
}

type SeatStatus int16

const (
	SeatAvailable SeatStatus = 1
	SeatHeld      SeatStatus = 2
	SeatSold      SeatStatus = 3
)

func (s SeatStatus) Name() string {
	switch s {
	case SeatAvailable:
		return "available"
	case SeatHeld:
		return "held"
	case SeatSold:
		return "sold"
	}
	return "unknown"
}

type Seat struct {
	ID           int64
	TicketTypeID int64
	Section      string
	Row          string
	Number       int
	Status       SeatStatus
	X, Y         *float64
	Accessible   bool
}

func (s Seat) Label() string {
	l := s.Row + "-" + strconv.Itoa(s.Number)
	if s.Section != "" {
		return s.Section + " " + l
	}
	return l
}

type PromoKind int16

const (
	PromoPercent PromoKind = 1
	PromoFixed   PromoKind = 2
)

type Promotion struct {
	ID         int64
	EventID    int64
	Code       string
	Kind       PromoKind
	Value      int64
	MaxUses    int
	Used       int
	MinTickets int
	ValidFrom  *time.Time
	ValidTo    *time.Time
	Active     bool
}

func NormalizeCode(c string) string { return strings.ToUpper(strings.TrimSpace(c)) }

func (p Promotion) UsableFor(now time.Time, tickets int) string {
	switch {
	case !p.Active:
		return "the promo code is not active"
	case p.ValidFrom != nil && now.Before(*p.ValidFrom):
		return "the promo code is not valid yet"
	case p.ValidTo != nil && now.After(*p.ValidTo):
		return "the promo code has expired"
	case p.MaxUses > 0 && p.Used >= p.MaxUses:
		return "the promo code has been fully used"
	case tickets < p.MinTickets:
		return "the promo code needs at least " + strconv.Itoa(p.MinTickets) + " tickets"
	}
	return ""
}

func (p Promotion) DiscountFor(subtotal int64) int64 {
	var d int64
	switch p.Kind {
	case PromoPercent:
		d = subtotal * p.Value / 100
	case PromoFixed:
		d = p.Value
	}
	return max(0, min(d, subtotal))
}

type EventReport struct {
	EventID     int64
	Orders      int
	TicketsSold int
	Revenue     int64
	CheckedIn   int
	Refunded    int64
	Invited     int
	ByType      []TypeSales
	Daily       []DailySales
}

type DailySales struct {
	Day     time.Time
	Orders  int
	Tickets int
	Revenue int64
}

type TypeSales struct {
	TicketTypeID int64
	Name         string
	Total        int
	Sold         int
	Held         int
	Revenue      int64
}

type SessionStatus int16

const (
	SessionScheduled SessionStatus = 1
	SessionCancelled SessionStatus = 2
)

func (s SessionStatus) Name() string {
	if s == SessionCancelled {
		return "cancelled"
	}
	return "scheduled"
}

type Session struct {
	ID           int64
	EventID      int64
	StartsAt     time.Time
	EndsAt       time.Time
	Label        string
	Status       SessionStatus
	CancelReason string
}

func (s Session) Over(now time.Time) bool { return !now.Before(s.EndsAt) }
