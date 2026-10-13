package model

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalid         = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
	ErrUnavailable     = errors.New("dependency unavailable")
	ErrUnauthenticated = errors.New("unauthenticated")
)

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }

func (e *ValidationError) Unwrap() error { return ErrInvalid }

func invalid(field, reason string) error {
	return &ValidationError{Field: field, Reason: reason}
}
