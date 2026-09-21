package port

import "errors"

var (
	ErrNotFound = errors.New("resource not found")

	ErrEmptyOrder                 = errors.New("order must contain at least one item")
	ErrInvalidStatusTransition    = errors.New("invalid order status transition")
	ErrOrderAlreadyTerminal       = errors.New("order is already in a terminal state")
	ErrRestaurantInactive         = errors.New("restaurant is not active")
	ErrMenuItemUnavailable        = errors.New("one or more menu items are unavailable")
	ErrPaymentAuthorizationFailed = errors.New("payment authorization failed")

	ErrDriverUnavailable    = errors.New("driver is not active")
	ErrAssignmentNotPending = errors.New("driver assignment is not pending")
)
