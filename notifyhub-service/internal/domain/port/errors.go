package port

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrQueueFull    = errors.New("worker queue full, try again later")
	ErrQueueClosed  = errors.New("worker queue closed")
)
