package common

import "errors"

var (
	ErrNotFound      = errors.New("record not found")
	ErrDBFailed      = errors.New("database operation failed")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInvalidStatus = errors.New("invalid status transition")
	ErrNoAdAvailable = errors.New("no ad available for this position")
)
