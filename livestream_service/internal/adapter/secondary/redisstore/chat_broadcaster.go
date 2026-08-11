package redisstore

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

func chatChannel(roomID string) string { return "chat:" + roomID }

type chatBroadcaster struct {
	client *redis.Client

	mu   sync.Mutex
	hubs map[string]*roomHub
}

func NewChatBroadcaster(client *redis.Client) port.ChatBroadcaster {
	return &chatBroadcaster{client: client, hubs: make(map[string]*roomHub)}
}

func (b *chatBroadcaster) Publish(ctx context.Context, msg model.ChatMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, chatChannel(msg.RoomID), payload).Err()
}

func (b *chatBroadcaster) Subscribe(ctx context.Context, roomID string) (<-chan model.ChatMessage, func(), error) {
	b.mu.Lock()
	hub, ok := b.hubs[roomID]
	if !ok {
		hub = newRoomHub(b.client, roomID)
		if err := hub.start(); err != nil {
			b.mu.Unlock()
			return nil, nil, err
		}
		b.hubs[roomID] = hub
	}
	ch, id := hub.addListener()
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		hub.removeListener(id)
		empty := hub.isEmpty()
		if empty {
			delete(b.hubs, roomID)
		}
		b.mu.Unlock()
		if empty {
			hub.stop()
		}
	}
	return ch, unsub, nil
}

type roomHub struct {
	client *redis.Client
	roomID string
	sub    *redis.PubSub
	cancel context.CancelFunc

	mu        sync.Mutex
	listeners map[int]chan model.ChatMessage
	nextID    int
}

func newRoomHub(client *redis.Client, roomID string) *roomHub {
	return &roomHub{client: client, roomID: roomID, listeners: make(map[int]chan model.ChatMessage)}
}

func (h *roomHub) start() error {
	ctx, cancel := context.WithCancel(context.Background())
	h.sub = h.client.Subscribe(ctx, chatChannel(h.roomID))
	if _, err := h.sub.Receive(ctx); err != nil {
		cancel()
		return err
	}
	h.cancel = cancel
	go h.readLoop(ctx)
	return nil
}

func (h *roomHub) readLoop(ctx context.Context) {
	ch := h.sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case redisMsg, ok := <-ch:
			if !ok {
				return
			}
			var msg model.ChatMessage
			if err := json.Unmarshal([]byte(redisMsg.Payload), &msg); err != nil {
				continue
			}
			h.broadcast(msg)
		}
	}
}

func (h *roomHub) broadcast(msg model.ChatMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.listeners {
		select {
		case ch <- msg:
		default:
			// Slow consumer: drop rather than block the whole room's fanout.
		}
	}
}

func (h *roomHub) addListener() (chan model.ChatMessage, int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextID
	h.nextID++
	ch := make(chan model.ChatMessage, 32)
	h.listeners[id] = ch
	return ch, id
}

func (h *roomHub) removeListener(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.listeners[id]; ok {
		delete(h.listeners, id)
		close(ch)
	}
}

func (h *roomHub) isEmpty() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.listeners) == 0
}

func (h *roomHub) stop() {
	h.cancel()
	_ = h.sub.Close()
}
