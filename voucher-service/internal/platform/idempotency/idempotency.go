package idempotency

import (
	"context"
	"time"
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Record struct {
	Key            string
	Status         Status
	ResponseStatus int
	ResponseBody   []byte
}

type Store interface {
	Claim(ctx context.Context, key, requestHash string, ttl time.Duration) (claimed bool, err error)
	Get(ctx context.Context, key string) (*Record, error)
	Complete(ctx context.Context, key string, responseStatus int, responseBody []byte) error
	Fail(ctx context.Context, key string) error
	Release(ctx context.Context, key string) error
}
