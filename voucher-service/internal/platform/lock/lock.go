package lock

import (
	"context"
	"errors"
	"time"
)

var ErrLockUnavailable = errors.New("lock backend unavailable")

type Locker interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (release func(context.Context) error, ok bool, err error)
}
