package common

import "errors"

// Domain errors — use errors.Is() for comparison, never string matching.
var (
	ErrNotFound = errors.New("record not found")
	ErrDBFailed = errors.New("database operation failed")
)
