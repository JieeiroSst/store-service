package model

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalid          = errors.New("invalid argument")
	ErrConflict         = errors.New("conflict")
	ErrUnavailable      = errors.New("dependency unavailable")
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrNoActiveCampaign = errors.New("no lucky wheel campaign is running")
	ErrNoSpinsLeft      = errors.New("no plays left")
	ErrAlreadyCheckedIn = errors.New("already checked in today")
	ErrRateLimited      = errors.New("too many requests")
	ErrBlocked          = errors.New("this account is restricted from lucky games")
	ErrDeviceLimit      = errors.New("too many accounts have played from this device")
	ErrDeviceRequired   = errors.New("X-Device-Id header is required")
	ErrCreditLimit      = errors.New("extra spin limit reached for today")
	ErrCampaignOverlap  = errors.New("another active campaign overlaps this time window")
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
