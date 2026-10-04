package redisstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	renewScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0`)
	releaseScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0`)
)

type Elector struct {
	rdb    goredis.UniversalClient
	key    string
	id     string
	ttl    time.Duration
	leader atomic.Bool
	cancel context.CancelFunc
	done   chan struct{}
}

func NewElector(rdb goredis.UniversalClient, prefix string, ttl time.Duration) *Elector {
	var b [8]byte
	_, _ = rand.Read(b[:])
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	return &Elector{rdb: rdb, key: prefix + "leader", id: hex.EncodeToString(b[:]), ttl: ttl}
}

func (e *Elector) IsLeader() bool { return e.leader.Load() }

func (e *Elector) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.done = make(chan struct{})
	go func() {
		defer close(e.done)
		t := time.NewTicker(e.ttl / 3)
		defer t.Stop()
		for {
			e.campaign(ctx)
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

func (e *Elector) campaign(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, e.ttl/3)
	defer cancel()
	if e.leader.Load() {
		n, err := renewScript.Run(ctx, e.rdb, []string{e.key}, e.id, e.ttl.Milliseconds()).Int()
		if err != nil || n == 0 {
			e.leader.Store(false)
			log.Printf("leader: lost leadership (%v)", err)
		}
		return
	}
	ok, err := e.rdb.SetNX(ctx, e.key, e.id, e.ttl).Result()
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("leader: %v", err)
		return
	}
	if ok {
		e.leader.Store(true)
		log.Printf("leader: acquired as %s", e.id)
	}
}

func (e *Elector) Stop() {
	if e.cancel == nil {
		return
	}
	e.cancel()
	<-e.done
	if e.leader.Swap(false) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = releaseScript.Run(ctx, e.rdb, []string{e.key}, e.id).Err()
	}
}
