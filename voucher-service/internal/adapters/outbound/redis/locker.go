package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/platform/lock"
	"github.com/redis/go-redis/v9"
)

const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`

const (
	acquireAttempts = 3
	acquireBackoff  = 30 * time.Millisecond
)

type Locker struct {
	client *redis.Client
}

func NewLocker(client *redis.Client) lock.Locker {
	return &Locker{client: client}
}

func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (func(context.Context) error, bool, error) {
	token, err := randomToken()
	if err != nil {
		return nil, false, err
	}

	var acquired bool
	var lastErr error
	for i := 0; i < acquireAttempts; i++ {
		ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
		if err != nil {
			lastErr = err
			break
		}
		if ok {
			acquired = true
			break
		}
		if i < acquireAttempts-1 {
			select {
			case <-time.After(acquireBackoff):
			case <-ctx.Done():
				return nil, false, ctx.Err()
			}
		}
	}
	if lastErr != nil {
		return nil, false, lastErr
	}
	if !acquired {
		return nil, false, nil
	}

	release := func(releaseCtx context.Context) error {
		return l.client.Eval(releaseCtx, releaseScript, []string{key}, token).Err()
	}
	return release, true, nil
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
