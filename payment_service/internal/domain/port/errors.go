package port

import "errors"

var (
	ErrNotFound             = errors.New("payment not found")
	ErrInvalidProvider      = errors.New("invalid or missing payment provider")
	ErrUnsupportedProvider  = errors.New("unsupported payment provider")
	ErrInvalidAmount        = errors.New("amount must be greater than zero")
	ErrInvalidCurrency      = errors.New("currency is required")
	ErrGatewayRequestFailed = errors.New("payment gateway request failed")
	ErrRefundNotSupported   = errors.New("refund is not supported by this provider")
	ErrPaymentNotRefundable = errors.New("payment is not in a refundable state")
	ErrInvalidRefundAmount  = errors.New("refund amount exceeds the payment's outstanding balance")
	ErrInvalidWebhook       = errors.New("invalid webhook signature or payload")
)
