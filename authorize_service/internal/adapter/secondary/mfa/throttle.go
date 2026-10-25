package mfa

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const throttlePrefix = "authorize_service:mfa_failures:"

var recordFailure = redis.NewScript(`
local n = redis.call("INCR", KEYS[1])
if n == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return n
`)

type Throttle struct {
	rdb         redis.UniversalClient
	maxFailures int64
	window      time.Duration
}

func NewThrottle(rdb redis.UniversalClient, maxFailures int64, window time.Duration) *Throttle {
	return &Throttle{rdb: rdb, maxFailures: maxFailures, window: window}
}

func (t *Throttle) Blocked(ctx context.Context, key string) (bool, error) {
	n, err := t.rdb.Get(ctx, throttlePrefix+key).Int64()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return n >= t.maxFailures, nil
}

func (t *Throttle) RecordFailure(ctx context.Context, key string) error {
	return recordFailure.Run(ctx, t.rdb, []string{throttlePrefix + key}, t.window.Milliseconds()).Err()
}

func (t *Throttle) Reset(ctx context.Context, key string) error {
	return t.rdb.Del(ctx, throttlePrefix+key).Err()
}
