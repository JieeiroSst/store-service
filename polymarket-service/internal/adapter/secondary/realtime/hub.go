package realtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/redis/go-redis/v9"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

const subscriberBuffer = 64

type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan port.StreamMessage]struct{}
}

func NewHub() *Hub { return &Hub{subs: map[string]map[chan port.StreamMessage]struct{}{}} }

func (h *Hub) Publish(topic string, payload any) {
	msg, ok := payload.(port.StreamMessage)
	if !ok {
		msg = port.StreamMessage{Type: "message", Payload: payload}
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[topic] {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (h *Hub) Subscribe(topic string) (<-chan port.StreamMessage, func()) {
	ch := make(chan port.StreamMessage, subscriberBuffer)
	h.mu.Lock()
	if h.subs[topic] == nil {
		h.subs[topic] = map[chan port.StreamMessage]struct{}{}
	}
	h.subs[topic][ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if _, ok := h.subs[topic][ch]; ok {
			delete(h.subs[topic], ch)
			if len(h.subs[topic]) == 0 {
				delete(h.subs, topic)
			}
			close(ch)
		}
	}
}

type Bus interface {
	port.Publisher
	port.EventStream
}

func NewBus(lc fx.Lifecycle, cfg *config.Config) (Bus, error) {
	if cfg.Redis.Addr == "" {
		return NewHub(), nil
	}
	client := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("realtime: cannot reach redis at %s: %w", cfg.Redis.Addr, err)
	}
	hub := NewRedisHub(client)
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return hub.Close() }})
	return hub, nil
}

var Module = fx.Options(
	fx.Provide(
		NewBus,
		func(b Bus) port.Publisher { return b },
		func(b Bus) port.EventStream { return b },
	),
)
