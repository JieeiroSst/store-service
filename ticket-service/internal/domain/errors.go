package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrInvalid             = errors.New("invalid input")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrSoldOut             = errors.New("sold out")
	ErrAlreadyCheckedIn    = errors.New("ticket already checked in")
	ErrUpstreamUnavailable = errors.New("upstream service unavailable")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrBusy                = errors.New("service is busy")
	ErrPaymentFailed       = errors.New("payment failed")
)

type SoldOutError struct {
	TicketTypeID int64
	Name         string
}

func (e *SoldOutError) Error() string {
	if e.Name == "" {
		return "sold out"
	}
	return "sold out: not enough " + e.Name + " tickets left"
}

func (e *SoldOutError) Is(target error) bool { return target == ErrSoldOut }
