package model

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalid         = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
	ErrUpstream        = errors.New("upstream unavailable")
	ErrPaymentFailed   = errors.New("payment failed")
	ErrUnauthenticated = errors.New("unauthenticated")
)

func Invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

func Conflict(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrConflict, fmt.Sprintf(format, args...))
}

func PaymentFailed(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrPaymentFailed, fmt.Sprintf(format, args...))
}
