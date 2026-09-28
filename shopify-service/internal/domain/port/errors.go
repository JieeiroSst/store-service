package port

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidWebhook       = errors.New("invalid webhook signature or payload")
	ErrShopifyRequestFailed = errors.New("shopify request failed")
	ErrShopifyUserError     = errors.New("shopify rejected the request")
)
