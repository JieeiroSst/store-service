package common

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrDBFailed       = errors.New("database operation failed")
	ErrInvalidRequest = errors.New("invalid request")
	ErrConflict       = errors.New("conflict")
	ErrOutOfStock     = errors.New("not enough stock")
	ErrPaymentFailed  = errors.New("payment failed")
	ErrUpstream       = errors.New("upstream service unavailable")
)
