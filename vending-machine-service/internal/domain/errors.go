package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrConflict           = errors.New("conflict")
	ErrMachineUnavailable = errors.New("machine is not active")
	ErrSessionNotActive   = errors.New("session is not active")
	ErrProductInactive    = errors.New("product is not active")
	ErrOutOfStock         = errors.New("slot is out of stock")
	ErrWrongMachine       = errors.New("slot belongs to a different machine")
	ErrReservationClosed  = errors.New("reservation is no longer pending")
	ErrPaymentFailed      = errors.New("payment failed")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrPriceChanged       = errors.New("product price changed, please retry")
	ErrCouponRejected     = errors.New("coupon rejected")
	ErrUpstream           = errors.New("upstream service unavailable")
	ErrPaymentPending     = errors.New("payment result is not known yet")
)

type PendingPaymentError struct {
	PaymentID string
}

func (e *PendingPaymentError) Error() string {
	return fmt.Sprintf("payment %s is being confirmed with the provider, check its status later", e.PaymentID)
}

func (e *PendingPaymentError) Unwrap() error { return ErrPaymentPending }
