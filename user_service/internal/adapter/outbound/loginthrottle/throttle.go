package loginthrottle

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "user_service:login_failures:"

var recordScript = redis.NewScript(`
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

func New(rdb redis.UniversalClient, maxFailures int64, window time.Duration) *Throttle {
	return &Throttle{rdb: rdb, maxFailures: maxFailures, window: window}
}

func key(username string) string {
	return keyPrefix + strings.ToLower(strings.TrimSpace(username))
}

func (t *Throttle) Blocked(ctx context.Context, username string) (bool, error) {
	n, err := t.rdb.Get(ctx, key(username)).Int64()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return n >= t.maxFailures, nil
}

func (t *Throttle) RecordFailure(ctx context.Context, username string) error {
	return recordScript.Run(ctx, t.rdb, []string{key(username)}, t.window.Milliseconds()).Err()
}

func (t *Throttle) Reset(ctx context.Context, username string) error {
	return t.rdb.Del(ctx, key(username)).Err()
}
