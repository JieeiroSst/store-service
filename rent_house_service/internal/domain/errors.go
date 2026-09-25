package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrInvalid             = errors.New("invalid input")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrDatesUnavailable    = errors.New("selected dates are not available")
	ErrUpstreamUnavailable = errors.New("upstream service unavailable")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrPaymentFailed       = errors.New("payment failed")
)
