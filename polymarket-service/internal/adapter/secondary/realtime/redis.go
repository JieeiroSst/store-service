package realtime

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

const (
	channelPrefix  = "polymarket:"
	publishTimeout = 500 * time.Millisecond
)

type RedisHub struct {
	client *redis.Client
	pubsub *redis.PubSub
	cancel context.CancelFunc

	mu   sync.Mutex
	subs map[string]map[chan port.StreamMessage]struct{}
}

type wire struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewRedisHub(client *redis.Client) *RedisHub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &RedisHub{
		client: client, pubsub: client.Subscribe(ctx), cancel: cancel,
		subs: map[string]map[chan port.StreamMessage]struct{}{},
	}
	go h.listen()
	return h
}

func (h *RedisHub) Publish(topic string, payload any) {
	msg, ok := payload.(port.StreamMessage)
	if !ok {
		msg = port.StreamMessage{Type: "message", Payload: payload}
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()
	_ = h.client.Publish(ctx, channelPrefix+topic, body).Err()
}

func (h *RedisHub) Subscribe(topic string) (<-chan port.StreamMessage, func()) {
	ch := make(chan port.StreamMessage, subscriberBuffer)

	h.mu.Lock()
	first := h.subs[topic] == nil
	if first {
		h.subs[topic] = map[chan port.StreamMessage]struct{}{}
	}
	h.subs[topic][ch] = struct{}{}
	h.mu.Unlock()
	if first {
		_ = h.pubsub.Subscribe(context.Background(), channelPrefix+topic)
	}

	return ch, func() {
		h.mu.Lock()
		if _, ok := h.subs[topic][ch]; !ok {
			h.mu.Unlock()
			return
		}
		delete(h.subs[topic], ch)
		close(ch)
		last := len(h.subs[topic]) == 0
		if last {
			delete(h.subs, topic)
		}
		h.mu.Unlock()
		if last {
			_ = h.pubsub.Unsubscribe(context.Background(), channelPrefix+topic)
		}
	}
}

func (h *RedisHub) listen() {
	for m := range h.pubsub.Channel() {
		var w wire
		if err := json.Unmarshal([]byte(m.Payload), &w); err != nil {
			continue
		}
		topic := strings.TrimPrefix(m.Channel, channelPrefix)
		h.mu.Lock()
		for ch := range h.subs[topic] {
			select {
			case ch <- port.StreamMessage{Type: w.Type, Payload: w.Payload}:
			default:
			}
		}
		h.mu.Unlock()
	}
}

func (h *RedisHub) Close() error {
	h.cancel()
	err := h.pubsub.Close()
	if cerr := h.client.Close(); err == nil {
		err = cerr
	}
	return err
}
