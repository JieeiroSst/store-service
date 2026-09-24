package common

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrDBFailed          = errors.New("database operation failed")
	ErrInvalidRequest    = errors.New("invalid request")
	ErrIntegrity         = errors.New("stored document failed its integrity check")
	ErrForbidden         = errors.New("operation not allowed")
	ErrTooLarge          = errors.New("file too large")
	ErrUnsupportedMedia  = errors.New("unsupported file type")
	ErrMalicious         = errors.New("file rejected as unsafe")
	ErrUpstream          = errors.New("upstream service failed")
	ErrInvalidSignature  = errors.New("invalid digital signature")
	ErrInvalidTransition = errors.New("invalid state transition")
)
