package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrDuplicatePAN = errors.New("pan already issued")
	ErrCVVMismatch  = errors.New("cvv does not match")
	ErrUnavailable  = errors.New("dependency unavailable")
)

type InvalidError struct{ Msg string }

func (e *InvalidError) Error() string { return e.Msg }

func Invalid(format string, args ...any) error {
	return &InvalidError{Msg: fmt.Sprintf(format, args...)}
}

func IsInvalid(err error) bool {
	var ie *InvalidError
	return errors.As(err, &ie)
}

type ConflictError struct{ Msg string }

func (e *ConflictError) Error() string { return e.Msg }

func Conflict(format string, args ...any) error {
	return &ConflictError{Msg: fmt.Sprintf(format, args...)}
}

func IsConflict(err error) bool {
	var ce *ConflictError
	return errors.As(err, &ce)
}
