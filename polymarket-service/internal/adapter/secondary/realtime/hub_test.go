package realtime

import (
	"testing"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func TestPublishReachesOnlyTopicSubscribers(t *testing.T) {
	h := NewHub()
	a, cancelA := h.Subscribe("market:1")
	b, cancelB := h.Subscribe("market:2")
	defer cancelB()

	h.Publish("market:1", port.StreamMessage{Type: "trade", Payload: 1})
	if msg := <-a; msg.Type != "trade" {
		t.Fatalf("got %+v", msg)
	}
	select {
	case msg := <-b:
		t.Fatalf("other topic received %+v", msg)
	default:
	}

	cancelA()
	cancelA() // idempotent
	h.Publish("market:1", port.StreamMessage{Type: "trade"})
}

func TestSlowSubscriberDoesNotBlockPublisher(t *testing.T) {
	h := NewHub()
	_, cancel := h.Subscribe("t")
	defer cancel()
	for i := 0; i < subscriberBuffer*3; i++ {
		h.Publish("t", port.StreamMessage{Type: "x"})
	}
}
