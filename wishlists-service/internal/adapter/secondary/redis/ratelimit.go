package rediscache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var incrScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return n
`)

type RateLimiter struct {
	rdb       *redis.Client
	prefix    string
	opTimeout time.Duration
	now       func() time.Time
}

func NewRateLimiter(rdb *redis.Client, prefix string, opTimeout time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, prefix: prefix, opTimeout: opTimeout, now: time.Now}
}

func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	bucket := l.now().UnixMilli() / window.Milliseconds()
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), l.opTimeout)
	defer cancel()
	n, err := incrScript.Run(c, l.rdb, []string{l.prefix + "ratelimit:" + key + ":" + strconv.FormatInt(bucket, 10)}, window.Milliseconds()).Int()
	if err != nil {
		return true, err
	}
	return n <= limit, nil
}

type NoopRateLimiter struct{}

func (NoopRateLimiter) Allow(context.Context, string, int, time.Duration) (bool, error) {
	return true, nil
}

type Health struct {
	rdb *redis.Client
}

func (h Health) Name() string { return "redis" }

func (h Health) Critical() bool { return false }

func (h Health) Ping(ctx context.Context) error { return h.rdb.Ping(ctx).Err() }
