package domain

import "time"

type OrderStatus int16

const (
	OrderPending   OrderStatus = 1
	OrderPaid      OrderStatus = 2
	OrderExpired   OrderStatus = 3
	OrderCancelled OrderStatus = 4
	OrderRefunded  OrderStatus = 5
)

func (s OrderStatus) Name() string {
	switch s {
	case OrderPending:
		return "pending"
	case OrderPaid:
		return "paid"
	case OrderExpired:
		return "expired"
	case OrderCancelled:
		return "cancelled"
	case OrderRefunded:
		return "refunded"
	}
	return "unknown"
}

func ParseOrderStatus(s string) OrderStatus {
	for _, st := range []OrderStatus{OrderPending, OrderPaid, OrderExpired, OrderCancelled, OrderRefunded} {
		if st.Name() == s {
			return st
		}
	}
	return 0
}

type Order struct {
	ID            int64
	UserID        int64
	EventID       int64
	Status        OrderStatus
	Currency      string
	Subtotal      int64
	Discount      int64
	Total         int64
	PromoID       int64
	PromoCode     string
	BuyerName     string
	BuyerEmail    string
	BuyerPhone    string
	RequestID     string
	ExpiresAt     time.Time
	PaymentMethod PaymentMethod
	PaymentRef    string
	PaidAt        *time.Time
	RefundAmount  int64
	StatusNote    string
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SessionID     int64
	Items         []OrderItem
	Tickets       []Ticket
}

func (o Order) Quantity() int {
	n := 0
	for _, it := range o.Items {
		n += it.Quantity
	}
	return n
}

type OrderItem struct {
	ID           int64
	OrderID      int64
	TicketTypeID int64
	Name         string
	Quantity     int
	UnitPrice    int64
	SeatIDs      []int64
}

// TicketStatus is the state of a ticket in its life:
//
//	valid ──check in──▶ used ──revert──▶ valid
//	  │  ├──offer──▶ transferring ──accept/decline/cancel/expire──▶ valid (a new holder and a new code after accept)
//	  │  └──event ends unused──▶ expired
//	  └──refund, cancel, revoke──▶ void
type TicketStatus int16

const (
	TicketValid        TicketStatus = 1
	TicketVoid         TicketStatus = 2
	TicketUsed         TicketStatus = 3
	TicketExpired      TicketStatus = 4
	TicketTransferring TicketStatus = 5
	TicketListed       TicketStatus = 6
)

func (s TicketStatus) Name() string {
	switch s {
	case TicketValid:
		return "valid"
	case TicketVoid:
		return "void"
	case TicketUsed:
		return "used"
	case TicketExpired:
		return "expired"
	case TicketTransferring:
		return "transferring"
	case TicketListed:
		return "listed"
	}
	return "unknown"
}

func ParseTicketStatus(s string) TicketStatus {
	for _, st := range []TicketStatus{TicketValid, TicketVoid, TicketUsed, TicketExpired, TicketTransferring, TicketListed} {
		if st.Name() == s {
			return st
		}
	}
	return 0
}

type Ticket struct {
	ID            int64
	OrderID       int64
	EventID       int64
	TicketTypeID  int64
	SeatID        int64
	SeatLabel     string
	TypeName      string
	Code          string
	Status        TicketStatus
	CheckedInAt   *time.Time
	CheckedInBy   int64
	HolderID      int64
	HolderName    string
	HolderEmail   string
	TransferCount int
	CreatedAt     time.Time
	EventTitle    string
	EventStartsAt time.Time
	EventEndsAt   time.Time
	SessionID     int64
	SessionLabel  string
	Venue         string
}

type Attendee struct {
	Ticket
	BuyerName  string
	BuyerEmail string
}
