package common

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrDBFailed       = errors.New("database operation failed")
	ErrInvalidRequest = errors.New("invalid request")
)
