package redis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JIeeiroSst/video-service/config"
	goredis "github.com/redis/go-redis/v9"
)

const (
	viewsHash     = "video:views"
	invalidateCh  = "video:invalidate"
	queueStream   = "video:transcode"
	queueGroup    = "transcoders"
	viewDedupeTTL = 30 * time.Minute
)

// Store implements port.ViewCounter, port.JobQueue and port.InvalidationBus
// on one Redis: view counts, the transcode queue (a Stream with a consumer
// group) and cache-invalidation pub/sub.
type Store struct {
	rdb       *goredis.Client
	consumer  string
	claimIdle time.Duration
	parallel  int
}

func NewClient(cfg *config.Config) *goredis.Client {
	return goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password})
}

func NewStore(rdb *goredis.Client, cfg *config.Config) *Store {
	host, _ := os.Hostname()
	return &Store{
		rdb:       rdb,
		consumer:  fmt.Sprintf("%s-%d", host, os.Getpid()),
		claimIdle: cfg.Worker.ClaimIdle,
		parallel:  max(cfg.Worker.Concurrency, 1),
	}
}

// --- ViewCounter ---

func (s *Store) Incr(ctx context.Context, id, viewer string) error {
	ok, err := s.rdb.SetNX(ctx, "view:"+id+":"+viewer, 1, viewDedupeTTL).Result()
	if err != nil || !ok {
		return err
	}
	return s.rdb.HIncrBy(ctx, viewsHash, id, 1).Err()
}

func (s *Store) Counts(ctx context.Context, ids []string) (map[string]int64, error) {
	out := make(map[string]int64, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	vals, err := s.rdb.HMGet(ctx, viewsHash, ids...).Result()
	if err != nil {
		return nil, err
	}
	for i, v := range vals {
		if str, ok := v.(string); ok {
			var n int64
			fmt.Sscan(str, &n)
			out[ids[i]] = n
		}
	}
	return out, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	return s.rdb.HDel(ctx, viewsHash, id).Err()
}

// --- InvalidationBus ---

func (s *Store) Publish(ctx context.Context, id string) error {
	return s.rdb.Publish(ctx, invalidateCh, id).Err()
}

func (s *Store) Subscribe(ctx context.Context, fn func(id string)) error {
	sub := s.rdb.Subscribe(ctx, invalidateCh)
	defer sub.Close()
	// go-redis reconnects and resubscribes on its own.
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-sub.Channel():
			if !ok {
				return nil
			}
			fn(msg.Payload)
		}
	}
}

// --- JobQueue ---

func (s *Store) Enqueue(ctx context.Context, id string) error {
	return s.rdb.XAdd(ctx, &goredis.XAddArgs{Stream: queueStream, Values: map[string]any{"id": id}}).Err()
}

func (s *Store) Consume(ctx context.Context, handle func(ctx context.Context, id string) error) error {
	err := s.rdb.XGroupCreateMkStream(ctx, queueStream, queueGroup, "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return err
	}

	errc := make(chan error, s.parallel)
	for i := 0; i < s.parallel; i++ {
		go func() { errc <- s.loop(ctx, handle) }()
	}
	var first error
	for i := 0; i < s.parallel; i++ {
		if err := <-errc; err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (s *Store) loop(ctx context.Context, handle func(context.Context, string) error) error {
	for ctx.Err() == nil {
		msgs, err := s.next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			time.Sleep(2 * time.Second) // Redis hiccup; keep trying
			continue
		}
		for _, m := range msgs {
			id, _ := m.Values["id"].(string)
			if err := handle(ctx, id); err != nil {
				// Leave it unacknowledged: after claimIdle another worker
				// (or this one) reclaims and retries it.
				continue
			}
			s.rdb.XAck(context.WithoutCancel(ctx), queueStream, queueGroup, m.ID)
			s.rdb.XDel(context.WithoutCancel(ctx), queueStream, m.ID)
		}
	}
	return nil
}

// next prefers jobs abandoned by dead workers, then blocks for new ones.
func (s *Store) next(ctx context.Context) ([]goredis.XMessage, error) {
	claimed, _, err := s.rdb.XAutoClaim(ctx, &goredis.XAutoClaimArgs{
		Stream: queueStream, Group: queueGroup, Consumer: s.consumer,
		MinIdle: s.claimIdle, Start: "0", Count: 1,
	}).Result()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return nil, err
	}
	if len(claimed) > 0 {
		return claimed, nil
	}

	streams, err := s.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group: queueGroup, Consumer: s.consumer,
		Streams: []string{queueStream, ">"}, Count: 1, Block: 5 * time.Second,
	}).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil || len(streams) == 0 {
		return nil, err
	}
	return streams[0].Messages, nil
}
