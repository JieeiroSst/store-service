package outbound

import (
	"context"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type HomestayRepository interface {
	Create(ctx context.Context, h domain.Homestay, actor int64) (domain.Homestay, error)
	Update(ctx context.Context, h domain.Homestay, actor int64) (domain.Homestay, error)
	Get(ctx context.Context, id int64) (domain.Homestay, error)
	List(ctx context.Context, f domain.HomestayFilter) ([]domain.Homestay, error)
	SetStatus(ctx context.Context, id int64, s domain.HomestayStatus, actor int64) error
	SetReview(ctx context.Context, id int64, s domain.HomestayStatus, note string, actor int64) error

	SetAvailability(ctx context.Context, homestayID int64, from, to time.Time, price string, status domain.AvailabilityStatus) error
	Availability(ctx context.Context, homestayID int64, from, to time.Time) ([]domain.Slot, error)

	SetRate(ctx context.Context, r domain.Rate, actor int64) error
	DeleteRate(ctx context.Context, homestayID int64, model domain.RentalModel) error // ErrNotFound if none
	Rates(ctx context.Context, homestayID int64) ([]domain.Rate, error)

	ListAmenities(ctx context.Context) ([]domain.Amenity, error)
	CreateAmenity(ctx context.Context, a domain.Amenity) (domain.Amenity, error)
}

type ReserveParams struct {
	UserID     int64
	HomestayID int64
	CheckIn    time.Time
	CheckOut   time.Time
	Guests     int
	Currency   string
	Note       string
	RequestID  string
	ExpiresAt  time.Time
}

type BookingRepository interface {
	Reserve(ctx context.Context, p ReserveParams) (domain.Booking, error)
	Cancel(ctx context.Context, id int64, actor int64) (domain.Booking, error)
	Get(ctx context.Context, id int64) (domain.Booking, error)
	List(ctx context.Context, userID, hostID int64, limit, offset int) ([]domain.Booking, error)

	SetPaymentAttempt(ctx context.Context, id int64, method domain.PaymentMethod, ref string) (domain.Booking, error)
	MarkPaid(ctx context.Context, id int64, method domain.PaymentMethod, ref string) (domain.Booking, error)
	Expire(ctx context.Context, id int64) (bool, error)
	ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Booking, error)
}

type CreateLeaseParams struct {
	UserID     int64
	HomestayID int64
	Model      domain.RentalModel
	Start, End time.Time
	Periods    int
	BillingDay int
	Currency   string
	Rent       string
	Note       string
	RequestID  string
	ExpiresAt  time.Time
	Schedule   []domain.InvoicePlan
}

type InvoiceFilter struct {
	Status    domain.InvoiceStatus // 0 = any
	DueBefore time.Time            // zero = no bound; combined with Status unpaid this lists what is overdue or due soon
	UserID    int64                // 0 = everyone
	HostID    int64                // only invoices of this owner's homestays; 0 = everyone's
	Limit     int
	Offset    int
}

type LeaseRepository interface {
	Create(ctx context.Context, p CreateLeaseParams) (domain.Lease, error)
	Get(ctx context.Context, id int64) (domain.Lease, error) // with invoices
	List(ctx context.Context, userID, hostID int64, limit, offset int) ([]domain.Lease, error)
	ListInvoices(ctx context.Context, f InvoiceFilter) ([]domain.Invoice, error)

	SetInvoiceAttempt(ctx context.Context, invoiceID int64, method domain.PaymentMethod, ref string) (domain.Invoice, error)
	MarkInvoicePaid(ctx context.Context, invoiceID int64, method domain.PaymentMethod, ref string) (domain.Invoice, error)
	Expire(ctx context.Context, id int64) (bool, error)
	ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Lease, error)
	Terminate(ctx context.Context, id int64, from time.Time) (domain.Lease, error)
	GetInvoice(ctx context.Context, id int64) (domain.Invoice, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, r domain.Review) (domain.Review, error)
	ListByHomestay(ctx context.Context, homestayID int64, limit, offset int) ([]domain.Review, error)
}

type WishlistRepository interface {
	Add(ctx context.Context, userID, homestayID int64) error
	Remove(ctx context.Context, userID, homestayID int64) error
	List(ctx context.Context, userID int64, limit, offset int) ([]domain.Homestay, error)
}
