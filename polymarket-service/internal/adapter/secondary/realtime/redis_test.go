package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newHub(t *testing.T, addr string) *RedisHub {
	t.Helper()
	h := NewRedisHub(redis.NewClient(&redis.Options{Addr: addr}))
	t.Cleanup(func() { _ = h.Close() })
	return h
}

func recv(t *testing.T, ch <-chan port.StreamMessage) port.StreamMessage {
	t.Helper()
	select {
	case m := <-ch:
		return m
	case <-time.After(2 * time.Second):
		t.Fatal("no message received")
		return port.StreamMessage{}
	}
}

// subscribed waits until Redis has registered the subscription, since SUBSCRIBE is asynchronous.
func subscribed(t *testing.T, r *miniredis.Miniredis, channel string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if r.PubSubNumSub(channel)[channel] > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("channel %s never subscribed", channel)
}

func TestEventsCrossReplicas(t *testing.T) {
	r := miniredis.RunT(t)
	replicaA, replicaB := newHub(t, r.Addr()), newHub(t, r.Addr())

	ch, cancel := replicaB.Subscribe("market:1")
	defer cancel()
	other, cancelOther := replicaB.Subscribe("market:2")
	defer cancelOther()
	subscribed(t, r, channelPrefix+"market:1")

	replicaA.Publish("market:1", port.StreamMessage{Type: "trade", Payload: map[string]int{"size": 5}})

	msg := recv(t, ch)
	if msg.Type != "trade" {
		t.Fatalf("%+v", msg)
	}
	var payload map[string]int
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil || payload["size"] != 5 {
		t.Fatalf("payload %s %v", msg.Payload, err)
	}
	select {
	case m := <-other:
		t.Fatalf("another market's subscriber got %+v", m)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestSubscriptionEndsWithTheLastListener(t *testing.T) {
	r := miniredis.RunT(t)
	hub := newHub(t, r.Addr())

	a, cancelA := hub.Subscribe("market:9")
	b, cancelB := hub.Subscribe("market:9")
	subscribed(t, r, channelPrefix+"market:9")
	hub.Publish("market:9", port.StreamMessage{Type: "book"})
	recv(t, a)
	recv(t, b)

	cancelA()
	cancelA() // idempotent
	hub.Publish("market:9", port.StreamMessage{Type: "book"})
	recv(t, b)

	cancelB()
	deadline := time.Now().Add(2 * time.Second)
	for r.PubSubNumSub(channelPrefix + "market:9")[channelPrefix+"market:9"] > 0 {
		if time.Now().After(deadline) {
			t.Fatal("redis subscription must be dropped when nobody listens")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestPublishSurvivesRedisBeingDown(t *testing.T) {
	r := miniredis.RunT(t)
	hub := newHub(t, r.Addr())
	r.Close()
	done := make(chan struct{})
	go func() {
		hub.Publish("market:1", port.StreamMessage{Type: "trade"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("a Redis outage must not block or fail trading")
	}
}

var _ = context.Background
