package common

import "errors"

var (
	ErrNotFound              = errors.New("record not found")
	ErrDBFailed              = errors.New("database operation failed")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrPaymentMethodRequired = errors.New("a payment method is required to charge this subscription")
	ErrGatewayUnavailable    = errors.New("payment gateway unavailable")
	ErrSubscriptionNotActive = errors.New("subscription is not active")
	ErrPaymentFailed         = errors.New("payment charge was declined")
)
