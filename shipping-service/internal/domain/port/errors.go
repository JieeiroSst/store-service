package port

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrConflict           = errors.New("conflict")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrNoRoute            = errors.New("no delivery route available")
	ErrCarrierRejected    = errors.New("carrier rejected the request")
	ErrCarrierUnavailable = errors.New("carrier unavailable")
	ErrUnauthorized       = errors.New("unauthorized")
)
