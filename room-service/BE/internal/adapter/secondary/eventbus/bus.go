package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

const channel = "room-service:events"

type Local interface {
	Deliver(model.Event)
}

type Bus struct {
	local Local
	rdb   *redis.Client
}

func New(lc fx.Lifecycle, cfg *config.Config, local Local) *Bus {
	b := &Bus{local: local}
	if cfg.Redis.Addr == "" {
		log.Println("eventbus: REDIS_ADDR unset, events stay on this replica")
		return b
	}

	b.rdb = redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password})
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			pingCtx, done := context.WithTimeout(startCtx, 10*time.Second)
			defer done()
			if err := b.rdb.Ping(pingCtx).Err(); err != nil {
				return fmt.Errorf("redis: %w", err)
			}
			sub := b.rdb.Subscribe(ctx, channel)
			if _, err := sub.Receive(pingCtx); err != nil {
				return fmt.Errorf("redis subscribe: %w", err)
			}
			go b.consume(sub)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return b.rdb.Close()
		},
	})
	return b
}

func (b *Bus) Publish(ctx context.Context, e model.Event) error {
	if b.rdb == nil {
		b.local.Deliver(e)
		return nil
	}
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return b.rdb.Publish(ctx, channel, payload).Err()
}

func (b *Bus) consume(sub *redis.PubSub) {
	defer sub.Close()
	for msg := range sub.Channel() {
		var e model.Event
		if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
			log.Printf("eventbus: bad event: %v", err)
			continue
		}
		b.local.Deliver(e)
	}
}
