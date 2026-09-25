package domain

import "time"

type BookingStatus int16

const (
	BookingConfirmed BookingStatus = 1 // paid
	BookingCancelled BookingStatus = 2
	BookingPending   BookingStatus = 3 // nights held, waiting for payment
)

type Booking struct {
	ID          int64
	UserID      int64
	HomestayID  int64
	CheckIn     time.Time
	CheckOut    time.Time
	Guests      int
	Status      BookingStatus
	Currency    string
	Subtotal    string
	Discount    string
	TotalAmount string
	Note        string
	RequestID   string
	CreatedAt   time.Time

	PaymentMethod PaymentMethod
	PaymentRef    string
	PaidAt        *time.Time
	ExpiresAt     *time.Time
	Version       int64
}

func (b Booking) Nights() int { return int(b.CheckOut.Sub(b.CheckIn).Hours() / 24) }
