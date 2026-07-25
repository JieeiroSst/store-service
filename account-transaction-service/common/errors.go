package common

import "errors"

// Domain errors — use errors.Is() for comparison, never string matching.
var (
	ErrNotFound               = errors.New("record not found")
	ErrDBFailed               = errors.New("database operation failed")
	ErrInvalidAmount          = errors.New("amount must be greater than zero")
	ErrSameAccount            = errors.New("sender and receiver accounts must differ")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)
