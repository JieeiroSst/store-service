package http

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const dateLayout = "2006-01-02"

func parseDate(field, v string) (time.Time, error) {
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s must be YYYY-MM-DD", domain.ErrInvalid, field)
	}
	return t, nil
}

type homestayRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        int      `json:"type"`
	Status      int16    `json:"status"`
	PhoneNumber string   `json:"phone_number"`
	Address     string   `json:"address"`
	WardID      int      `json:"ward_id"`
	DistrictID  int      `json:"district_id"`
	ProvinceID  int      `json:"province_id"`
	Images      []string `json:"images"`
	Guests      int      `json:"guests"`
	Bedrooms    int      `json:"bedrooms"`
	Bathrooms   int      `json:"bathrooms"`
	AmenityIDs  []int    `json:"amenity_ids"`
	WalletID    *string  `json:"wallet_id"`
}

type homestayResponse struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Type          int            `json:"type"` // 1 homestay, 2 hotel, 3 rental house
	TypeName      string         `json:"type_name"`
	Status        int16          `json:"status"` // 1 active, 2 inactive, 3 pending approval, 4 rejected
	StatusName    string         `json:"status_name"`
	ReviewNote    string         `json:"review_note,omitempty"`
	PhoneNumber   string         `json:"phone_number"`
	Address       string         `json:"address"`
	WardID        int            `json:"ward_id"`
	DistrictID    int            `json:"district_id"`
	ProvinceID    int            `json:"province_id"`
	Images        []string       `json:"images"`
	Guests        int            `json:"guests"`
	Bedrooms      int            `json:"bedrooms"`
	Bathrooms     int            `json:"bathrooms"`
	AmenityIDs    []int          `json:"amenity_ids"`
	Version       int64          `json:"version"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	HostID        int64          `json:"host_id"`
	AcceptsWallet bool           `json:"accepts_wallet_payment"`
	Rating        float64        `json:"rating"`
	ReviewCount   int            `json:"review_count"`
	Rates         []rateResponse `json:"rates"`
}

func toHomestayResponse(h domain.Homestay) homestayResponse {
	return homestayResponse{
		ID: h.ID, Name: h.Name, Description: h.Description, Type: h.Type, TypeName: domain.PropertyType(h.Type).Name(),
		Status: int16(h.Status), StatusName: h.Status.Name(), ReviewNote: h.ReviewNote,
		PhoneNumber: h.PhoneNumber, Address: h.Address, WardID: h.WardID, DistrictID: h.DistrictID,
		ProvinceID: h.ProvinceID, Images: h.Images, Guests: h.Guests, Bedrooms: h.Bedrooms,
		Bathrooms: h.Bathrooms, AmenityIDs: h.AmenityIDs, Version: h.Version,
		CreatedAt: h.CreatedAt, UpdatedAt: h.UpdatedAt,
		HostID: h.HostID, AcceptsWallet: h.WalletID != "", Rating: math.Round(h.Rating*100) / 100, ReviewCount: h.ReviewCount, Rates: toRateResponses(h.Rates),
	}
}

type availabilityRequest struct {
	From   string      `json:"from"`
	To     string      `json:"to"` // exclusive
	Price  json.Number `json:"price"`
	Status int16       `json:"status"` // 1 available, 3 blocked
}

type slotResponse struct {
	Date   string `json:"date"`
	Price  string `json:"price"`
	Status int16  `json:"status"`
}

type amenityRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type amenityResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type bookingRequest struct {
	HomestayID int64  `json:"homestay_id"`
	CheckIn    string `json:"checkin_date"`
	CheckOut   string `json:"checkout_date"`
	Guests     int    `json:"guests"`
	Currency   string `json:"currency"`
	Note       string `json:"note"`
	RequestID  string `json:"request_id"`
}

type bookingResponse struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	HomestayID    int64      `json:"homestay_id"`
	CheckIn       string     `json:"checkin_date"`
	CheckOut      string     `json:"checkout_date"`
	Nights        int        `json:"nights"`
	Guests        int        `json:"guests"`
	Status        int16      `json:"status"`
	Currency      string     `json:"currency"`
	Subtotal      string     `json:"subtotal"`
	Discount      string     `json:"discount"`
	TotalAmount   string     `json:"total_amount"`
	Note          string     `json:"note"`
	RequestID     string     `json:"request_id"`
	CreatedAt     time.Time  `json:"created_at"`
	PaymentMethod string     `json:"payment_method,omitempty"`
	PaymentRef    string     `json:"payment_ref,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

func toBookingResponse(b domain.Booking) bookingResponse {
	return bookingResponse{
		ID: b.ID, UserID: b.UserID, HomestayID: b.HomestayID,
		CheckIn: b.CheckIn.Format(dateLayout), CheckOut: b.CheckOut.Format(dateLayout), Nights: b.Nights(),
		Guests: b.Guests, Status: int16(b.Status), Currency: b.Currency, Subtotal: b.Subtotal,
		Discount: b.Discount, TotalAmount: b.TotalAmount, Note: b.Note, RequestID: b.RequestID, CreatedAt: b.CreatedAt,
		PaymentMethod: string(b.PaymentMethod), PaymentRef: b.PaymentRef, PaidAt: b.PaidAt, ExpiresAt: pendingExpiry(b),
	}
}

func pendingExpiry(b domain.Booking) *time.Time {
	if b.Status == domain.BookingPending {
		return b.ExpiresAt
	}
	return nil
}

type payRequest struct {
	Method   string `json:"method"`   // "wallet" or "gateway"
	Provider string `json:"provider"` // gateway only: paypal, stripe, ...
}

type walletRequest struct {
	Currency string `json:"currency"`
}

type walletResponse struct {
	ID       string `json:"wallet_id"`
	Balance  int64  `json:"balance"` // minor units
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

type walletTxnResponse struct {
	ID          string    `json:"transaction_id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	ReferenceID string    `json:"reference_id,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// homestayPageResponse is one page of a search. To get the next page, repeat the same request
// with cursor set to next_cursor; it is absent on the last page.
type homestayPageResponse struct {
	Items      []homestayResponse `json:"items"`
	HasMore    bool               `json:"has_more"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

type rateRequest struct {
	Price      json.Number `json:"price"`
	Currency   string      `json:"currency"`
	MinPeriods int         `json:"min_periods"`
	Active     *bool       `json:"active"` // default true
}

type rateResponse struct {
	Model      string `json:"model"` // day, week, month, year
	Price      string `json:"price"` // per night / week / month / year
	Currency   string `json:"currency"`
	MinPeriods int    `json:"min_periods"`
	Active     bool   `json:"active"`
}

func toRateResponses(rs []domain.Rate) []rateResponse {
	out := make([]rateResponse, len(rs))
	for i, r := range rs {
		out[i] = rateResponse{Model: string(r.Model), Price: r.Price, Currency: r.Currency, MinPeriods: r.MinPeriods, Active: r.Active}
	}
	return out
}

type leaseRequest struct {
	HomestayID int64  `json:"homestay_id"`
	Model      string `json:"model"` // week, month, year
	StartDate  string `json:"start_date"`
	Periods    int    `json:"periods"`
	// BillingDay: day of month rent is collected (month, 1-28) or weekday 1=Mon..7=Sun (week).
	BillingDay int    `json:"billing_day"`
	Note       string `json:"note"`
	RequestID  string `json:"request_id"`
}

type invoiceResponse struct {
	ID            int64      `json:"id"`
	LeaseID       int64      `json:"lease_id"`
	UserID        int64      `json:"user_id,omitempty"`
	HomestayID    int64      `json:"homestay_id,omitempty"`
	Period        int        `json:"period"` // 1-based
	PeriodStart   string     `json:"period_start"`
	PeriodEnd     string     `json:"period_end"`
	DueDate       string     `json:"due_date"`
	Amount        string     `json:"amount"`
	Status        string     `json:"status"` // unpaid, paid, void
	Overdue       bool       `json:"overdue"`
	PaymentMethod string     `json:"payment_method,omitempty"`
	PaymentRef    string     `json:"payment_ref,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}

type leaseResponse struct {
	ID         int64             `json:"id"`
	UserID     int64             `json:"user_id"`
	HomestayID int64             `json:"homestay_id"`
	Model      string            `json:"model"`
	StartDate  string            `json:"start_date"`
	EndDate    string            `json:"end_date"`
	Periods    int               `json:"periods"`
	BillingDay int               `json:"billing_day"`
	Currency   string            `json:"currency"`
	Rent       string            `json:"rent"` // per period
	Status     string            `json:"status"`
	Note       string            `json:"note,omitempty"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"` // hold on the nights while unpaid
	CreatedAt  time.Time         `json:"created_at"`
	Invoices   []invoiceResponse `json:"invoices,omitempty"`
}

func toInvoiceResponse(i domain.Invoice, today time.Time) invoiceResponse {
	st := map[domain.InvoiceStatus]string{domain.InvoiceUnpaid: "unpaid", domain.InvoicePaid: "paid", domain.InvoiceVoid: "void", domain.InvoiceRefunded: "refunded"}[i.Status]
	return invoiceResponse{
		ID: i.ID, LeaseID: i.LeaseID, UserID: i.UserID, HomestayID: i.HomestayID, Period: i.PeriodNo + 1,
		PeriodStart: i.PeriodStart.Format(dateLayout), PeriodEnd: i.PeriodEnd.Format(dateLayout), DueDate: i.DueDate.Format(dateLayout),
		Amount: i.Amount, Status: st, Overdue: i.Overdue(today),
		PaymentMethod: string(i.PaymentMethod), PaymentRef: i.PaymentRef, PaidAt: i.PaidAt,
	}
}

func toLeaseResponse(l domain.Lease, today time.Time) leaseResponse {
	st := map[domain.LeaseStatus]string{domain.LeasePending: "pending", domain.LeaseActive: "active", domain.LeaseEnded: "ended", domain.LeaseCancelled: "cancelled"}[l.Status]
	r := leaseResponse{
		ID: l.ID, UserID: l.UserID, HomestayID: l.HomestayID, Model: string(l.Model),
		StartDate: l.StartDate.Format(dateLayout), EndDate: l.EndDate.Format(dateLayout), Periods: l.Periods,
		BillingDay: l.BillingDay, Currency: l.Currency, Rent: l.Rent, Status: st, Note: l.Note, CreatedAt: l.CreatedAt,
	}
	if l.Status == domain.LeasePending {
		r.ExpiresAt = l.ExpiresAt
	}
	for _, i := range l.Invoices {
		r.Invoices = append(r.Invoices, toInvoiceResponse(i, today))
	}
	return r
}

type reviewRequest struct {
	BookingID int64  `json:"booking_id"` // exactly one of booking_id / lease_id
	LeaseID   int64  `json:"lease_id"`
	Rating    int    `json:"rating"` // 1-5
	Comment   string `json:"comment"`
}

type reviewResponse struct {
	ID         int64     `json:"id"`
	HomestayID int64     `json:"homestay_id"`
	UserID     int64     `json:"user_id"`
	BookingID  int64     `json:"booking_id,omitempty"`
	LeaseID    int64     `json:"lease_id,omitempty"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type loyaltyResponse struct {
	Points   int64  `json:"points"`
	Tier     int    `json:"tier"`
	TierName string `json:"tier_name"`
}

type rejectRequest struct {
	Reason string `json:"reason"`
}
