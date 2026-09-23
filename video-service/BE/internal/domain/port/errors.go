package port

import "errors"

var (
	ErrNotFound = errors.New("video not found")
	ErrInvalid  = errors.New("invalid request")
)
