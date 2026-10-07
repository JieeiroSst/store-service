package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("this payment authentication belongs to another user")
	ErrExpired      = errors.New("payment authentication expired")
	ErrUnavailable  = errors.New("dependency unavailable")
	ErrOTPExpired   = errors.New("otp expired, request a new code")
	ErrOTPLimit     = errors.New("too many otp codes requested for this user, try again later")
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

type OTPMismatchError struct{ Remaining int }

func (e *OTPMismatchError) Error() string {
	if e.Remaining == 0 {
		return "incorrect otp, no attempts left: the payment authentication failed"
	}
	return fmt.Sprintf("incorrect otp, %d attempt(s) left", e.Remaining)
}

func IsOTPMismatch(err error) bool {
	var oe *OTPMismatchError
	return errors.As(err, &oe)
}
