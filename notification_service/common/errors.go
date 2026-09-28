package common

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrDBFailed       = errors.New("database operation failed")
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotConfigured  = errors.New("channel not configured")
	ErrNoActiveDevice = errors.New("user has no active device")
	ErrPermanent      = errors.New("permanent delivery failure")
	ErrInvalidToken   = errors.New("fcm rejected the device token")
	ErrUnknownContact = errors.New("no user is registered for this email or phone")
)
