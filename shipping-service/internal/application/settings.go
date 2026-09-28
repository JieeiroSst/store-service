package application

import (
	"fmt"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type Settings struct {
	MaxCODAmount      int64
	PlacementLease    time.Duration
	QuoteConcurrency  int
	OutboxBatch       int
	OutboxLease       time.Duration
	OutboxMaxAttempts int
}

func DefaultSettings() Settings {
	return Settings{
		MaxCODAmount:      50_000_000,
		PlacementLease:    30 * time.Second,
		QuoteConcurrency:  8,
		OutboxBatch:       20,
		OutboxLease:       time.Minute,
		OutboxMaxAttempts: 10,
	}
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", port.ErrInvalidInput, fmt.Sprintf(format, args...))
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

func normalizePage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
