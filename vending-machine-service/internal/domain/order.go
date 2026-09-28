package domain

import (
	"fmt"
	"time"
)

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "pending"
	ReservationConfirmed ReservationStatus = "confirmed"
	ReservationCancelled ReservationStatus = "cancelled"
	ReservationExpired   ReservationStatus = "expired"
)

type Reservation struct {
	ID          string
	SessionID   string
	InventoryID string
	ExpiresAt   time.Time
	Status      ReservationStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Reservation) Release(status ReservationStatus) error {
	if r.Status != ReservationPending {
		return ErrReservationClosed
	}
	r.Status = status
	return nil
}

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentCard   PaymentMethod = "card"
	PaymentCash   PaymentMethod = "cash"
	PaymentMobile PaymentMethod = "mobile"
	PaymentWallet PaymentMethod = "wallet"
)

func ParsePaymentMethod(s string) (PaymentMethod, error) {
	switch m := PaymentMethod(s); m {
	case PaymentCard, PaymentCash, PaymentMobile, PaymentWallet:
		return m, nil
	}
	return "", fmt.Errorf("%w: unknown payment method %q", ErrInvalidInput, s)
}

type Payment struct {
	ID            string
	SessionID     string
	ReservationID string
	AmountCents   int
	DiscountCents int
	CouponCode    string
	Currency      string
	Method        PaymentMethod
	Status        PaymentStatus
	TransactionID string
	Metadata      map[string]string
	CompletedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderProcessing OrderStatus = "processing"
	OrderCompleted  OrderStatus = "completed"
	OrderFailed     OrderStatus = "failed"
	OrderCancelled  OrderStatus = "cancelled"
	OrderRefunded   OrderStatus = "refunded"
)

type Order struct {
	ID            string
	OrderNo       int64
	SessionID     string
	ReservationID string
	PaymentID     string
	Status        OrderStatus
	FulfilledAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (o *Order) Refund(at time.Time) error {
	if o.Status != OrderCompleted {
		return ErrInvalidTransition
	}
	o.Status = OrderRefunded
	o.UpdatedAt = at
	return nil
}

func (o *Order) Fulfil(dispensed bool, at time.Time) error {
	if o.Status != OrderProcessing {
		return ErrInvalidTransition
	}
	if dispensed {
		o.Status = OrderCompleted
		o.FulfilledAt = &at
	} else {
		o.Status = OrderFailed
	}
	return nil
}

type Checkout struct {
	Order   *Order
	Payment *Payment
}
