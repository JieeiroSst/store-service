package order

import "errors"

var (
	ErrInvalidOrderTransition = errors.New("invalid order status transition")
	ErrEmptyOrder             = errors.New("order has no items")
	ErrVersionConflict        = errors.New("order version conflict")
	ErrOrderNotFound          = errors.New("order not found")
	ErrInvalidOrder           = errors.New("invalid order")
)
