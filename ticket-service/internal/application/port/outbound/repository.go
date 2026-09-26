package outbound

import (
	"context"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type DuplicateSpec struct {
	Title          string
	StartsAt       time.Time
	EndsAt         time.Time
	CopyPromotions bool
}

type EventRepository interface {
	Create(ctx context.Context, e domain.Event) (domain.Event, error)
	Update(ctx context.Context, e domain.Event) (domain.Event, error)
	Get(ctx context.Context, id int64) (domain.Event, error)
	SetStatus(ctx context.Context, id int64, from []domain.EventStatus, to domain.EventStatus, note string) (domain.Event, error)
	SetFeatured(ctx context.Context, id int64, featured bool) error

	Session(ctx context.Context, id int64) (domain.Session, error)
	CreateSession(ctx context.Context, s domain.Session, copyFrom int64) (domain.Session, error)
	UpdateSession(ctx context.Context, s domain.Session) (domain.Session, error)
	DeleteSession(ctx context.Context, eventID, sessionID int64) error
	CancelSession(ctx context.Context, eventID, sessionID int64, reason string) (domain.Session, error)
	Duplicate(ctx context.Context, srcID int64, spec DuplicateSpec) (domain.Event, error)
	Series(ctx context.Context, eventID int64) ([]domain.Event, error)
	Search(ctx context.Context, f domain.EventFilter) ([]domain.Event, error)
	ListByOrganizer(ctx context.Context, organizerID, afterID int64, limit int) ([]domain.Event, error)
	ListByStatus(ctx context.Context, status domain.EventStatus, afterID int64, limit int) ([]domain.Event, error)

	CreateTicketType(ctx context.Context, t domain.TicketType) (domain.TicketType, error)
	UpdateTicketType(ctx context.Context, t domain.TicketType) (domain.TicketType, error)
	TicketType(ctx context.Context, id int64) (domain.TicketType, error)
	AddSeats(ctx context.Context, typeID int64, seats []domain.Seat) (int, error)
	Seats(ctx context.Context, typeID int64) ([]domain.Seat, error)
	SetSeatMap(ctx context.Context, typeID int64, l domain.SeatMapLayout, seats []domain.SeatPosition) (int, error)
	SeatLayout(ctx context.Context, typeID int64) (domain.SeatMapLayout, error)
	ApplyVenueMap(ctx context.Context, eventID, venueID int64, drawing domain.SeatMapLayout, groups []domain.SeatGroup, replace bool) (int, error)
	EventSeatMap(ctx context.Context, eventID int64) (domain.EventSeatMap, error)

	Promotions(ctx context.Context, eventID int64) ([]domain.Promotion, error)
	PromotionByCode(ctx context.Context, eventID int64, code string) (domain.Promotion, error)
	CreatePromotion(ctx context.Context, p domain.Promotion) (domain.Promotion, error)
	UpdatePromotion(ctx context.Context, p domain.Promotion) (domain.Promotion, error)

	Report(ctx context.Context, eventID int64) (domain.EventReport, error)
}

type ReserveItem struct {
	TicketTypeID int64
	Quantity     int
	SeatIDs      []int64
}

type ReserveParams struct {
	UserID     int64
	EventID    int64
	Items      []ReserveItem
	Promo      *domain.Promotion
	BuyerName  string
	BuyerEmail string
	BuyerPhone string
	RequestID  string
	ExpiresAt  time.Time
	MaxPending int
	Comp       bool
	InvitedBy  int64
}

type OrderFilter struct {
	UserID  int64
	EventID int64
	Status  domain.OrderStatus
	AfterID int64
	Limit   int
}

type TransitionParams struct {
	ID     int64
	From   []domain.OrderStatus
	To     domain.OrderStatus // expired, cancelled or refunded
	Note   string
	Refund int64
	Force  bool
}

type OrderRepository interface {
	Reserve(ctx context.Context, p ReserveParams) (domain.Order, error)
	Get(ctx context.Context, id int64) (domain.Order, error)
	RequestIDExists(ctx context.Context, requestID string) (bool, error)
	FindByRequestID(ctx context.Context, userID int64, requestID string) (domain.Order, bool, error)
	Available(ctx context.Context, typeIDs []int64) (map[int64]int, error)
	List(ctx context.Context, f OrderFilter) ([]domain.Order, error)
	SetPaymentAttempt(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Order, error)
	MarkPaid(ctx context.Context, id int64, m domain.PaymentMethod, ref string, codes []string) (domain.Order, error)
	Transition(ctx context.Context, p TransitionParams) (domain.Order, error)
	ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Order, error)
	ListOpenOfCancelledEvents(ctx context.Context, limit int) ([]domain.Order, error)
	BuyersOfSession(ctx context.Context, sessionID int64, limit int) ([]int64, error)
	DueForReminder(ctx context.Context, from, to time.Time, limit int) ([]domain.Order, error)
	MarkReminded(ctx context.Context, id int64) (bool, error)
}

type TicketFilter struct {
	HolderID int64
	EventID  int64
	Status   domain.TicketStatus
	AfterID  int64
	Limit    int
}

type OfferParams struct {
	TicketID     int64
	FromUser     int64
	ToUser       int64
	ToEmail      string
	Message      string
	ExpiresAt    time.Time
	MaxTransfers int
}

type TicketRepository interface {
	Get(ctx context.Context, id int64) (domain.Ticket, error)
	ByCode(ctx context.Context, eventID int64, code string) (domain.Ticket, error)
	List(ctx context.Context, f TicketFilter) ([]domain.Ticket, error)
	History(ctx context.Context, id int64) ([]domain.TicketEvent, error)
	Attendees(ctx context.Context, eventID, afterID int64, limit int) ([]domain.Attendee, error)

	CheckIn(ctx context.Context, eventID int64, code string, by int64, at time.Time) (domain.Ticket, error)
	RevertCheckIn(ctx context.Context, id, by int64) (domain.Ticket, error)
	SetHolder(ctx context.Context, id int64, name, email string, by int64) (domain.Ticket, error)
	Reissue(ctx context.Context, id int64, newCode string, by int64) (domain.Ticket, error)
	Void(ctx context.Context, id, by int64, note string) (domain.Ticket, error)

	Offer(ctx context.Context, p OfferParams) (domain.Transfer, error)
	GetTransfer(ctx context.Context, id int64) (domain.Transfer, error)
	Incoming(ctx context.Context, userID int64, email string, afterID int64, limit int) ([]domain.Transfer, error)
	Outgoing(ctx context.Context, userID, afterID int64, limit int) ([]domain.Transfer, error)
	AcceptTransfer(ctx context.Context, id, newHolder int64, name, email, newCode string) (domain.Ticket, error)
	CloseTransfer(ctx context.Context, id int64, as domain.TransferStatus, by int64) (domain.Transfer, error)

	ExpireTickets(ctx context.Context, now time.Time, limit int) (int, error)
	ExpireTransfers(ctx context.Context, now time.Time, limit int) (int, error)
}

type StaffRepository interface {
	Add(ctx context.Context, eventID, userID, addedBy int64) error
	Remove(ctx context.Context, eventID, userID int64) error
	List(ctx context.Context, eventID int64) ([]domain.StaffMember, error)
	Has(ctx context.Context, eventID, userID int64) (bool, error)
	EventsFor(ctx context.Context, userID int64) ([]domain.Event, error)
}

type WaitlistRepository interface {
	Join(ctx context.Context, typeID, userID int64) error
	Leave(ctx context.Context, typeID, userID int64) error
	Mine(ctx context.Context, userID int64) ([]domain.WaitlistEntry, error)
	Pop(ctx context.Context, typeID int64, n int) ([]int64, error)
}

type WishlistRepository interface {
	Add(ctx context.Context, userID, eventID int64) error
	Remove(ctx context.Context, userID, eventID int64) error
	List(ctx context.Context, userID int64, limit, offset int) ([]domain.Event, error)
}

type NotificationRepository interface {
	Add(ctx context.Context, n domain.Notification) error
	List(ctx context.Context, userID int64, unreadOnly bool, afterID int64, limit int) ([]domain.Notification, error)
	UnreadCount(ctx context.Context, userID int64) (int, error)
	MarkRead(ctx context.Context, userID, id int64) error
	MarkAllRead(ctx context.Context, userID int64) (int, error)
	ClaimUnpushed(ctx context.Context, limit int, lease time.Duration) ([]domain.Notification, error)
	MarkPushed(ctx context.Context, id int64) error
	ClaimUnemailed(ctx context.Context, limit int, lease time.Duration) ([]domain.Notification, error)
	MarkEmailed(ctx context.Context, id int64) error
}

type ResaleFilter struct {
	EventID       int64
	TypeID        int64
	SellerID      int64
	Status        domain.ResaleStatus
	CheapestFirst bool
	AfterID       int64
	Limit         int
}

type CreateListing struct {
	TicketID     int64
	SellerID     int64
	Price        int64
	MaxTransfers int
}

type ResaleRepository interface {
	Create(ctx context.Context, p CreateListing) (domain.ResaleListing, error)
	Get(ctx context.Context, id int64) (domain.ResaleListing, error)
	List(ctx context.Context, f ResaleFilter) ([]domain.ResaleListing, error)
	Cancel(ctx context.Context, id, sellerID int64) (domain.ResaleListing, error)
	Claim(ctx context.Context, id, buyerID int64) (domain.ResaleListing, error)
	Release(ctx context.Context, id int64) error
	SetPayment(ctx context.Context, id int64, ref string, fee int64) error
	Complete(ctx context.Context, id, buyerID int64, email, newCode string) (domain.Ticket, error)
	Stale(ctx context.Context, before time.Time, limit int) ([]domain.ResaleListing, error)
	ExpireStarted(ctx context.Context, now time.Time, limit int) (int, error)
}

type VenueFilter struct {
	OwnerID int64
	Shared  bool
	AfterID int64
	Limit   int
}

type VenueRepository interface {
	Create(ctx context.Context, v domain.Venue) (domain.Venue, error)
	Get(ctx context.Context, id int64) (domain.Venue, error)
	Update(ctx context.Context, v domain.Venue) (domain.Venue, error)
	Delete(ctx context.Context, id int64) error
	SetShared(ctx context.Context, id int64, shared bool) error
	List(ctx context.Context, f VenueFilter) ([]domain.Venue, error)
}

type DocumentRepository interface {
	Save(ctx context.Context, d domain.OrderDocument) (bool, error)
	List(ctx context.Context, orderID int64) ([]domain.OrderDocument, error)
	Get(ctx context.Context, orderID int64, kind string) (domain.OrderDocument, error)
	MarkDone(ctx context.Context, orderID int64) error
	Pending(ctx context.Context, limit int) ([]int64, error)
}
