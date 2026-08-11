package redisstore

import (
	"context"
	"testing"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func TestChatBroadcasterSharesOneRedisSubscriptionPerRoom(t *testing.T) {
	b := NewChatBroadcaster(newTestRedisClient(t)).(*chatBroadcaster)
	ctx := context.Background()

	ch1, unsub1, err := b.Subscribe(ctx, "room-1")
	if err != nil {
		t.Fatalf("first Subscribe() error = %v", err)
	}
	ch2, unsub2, err := b.Subscribe(ctx, "room-1")
	if err != nil {
		t.Fatalf("second Subscribe() error = %v", err)
	}

	b.mu.Lock()
	hubCount := len(b.hubs)
	hub := b.hubs["room-1"]
	b.mu.Unlock()
	if hubCount != 1 {
		t.Fatalf("expected exactly one hub for room-1 shared by both subscribers, got %d", hubCount)
	}

	hub.mu.Lock()
	listenerCount := len(hub.listeners)
	hub.mu.Unlock()
	if listenerCount != 2 {
		t.Errorf("expected 2 local listeners on the shared hub, got %d", listenerCount)
	}

	msg := model.ChatMessage{RoomID: "room-1", Body: "hello"}
	if err := b.Publish(ctx, msg); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	for i, ch := range []<-chan model.ChatMessage{ch1, ch2} {
		select {
		case got := <-ch:
			if got.Body != "hello" {
				t.Errorf("listener %d got body %q, want %q", i, got.Body, "hello")
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("listener %d: timed out waiting for the fanned-out message", i)
		}
	}

	unsub1()
	b.mu.Lock()
	_, stillPresent := b.hubs["room-1"]
	b.mu.Unlock()
	if !stillPresent {
		t.Error("hub should still exist while one listener remains subscribed")
	}

	unsub2()
	b.mu.Lock()
	_, present := b.hubs["room-1"]
	b.mu.Unlock()
	if present {
		t.Error("hub should be torn down once the last local listener unsubscribes")
	}
}

func TestChatBroadcasterIndependentRoomsGetIndependentHubs(t *testing.T) {
	b := NewChatBroadcaster(newTestRedisClient(t)).(*chatBroadcaster)
	ctx := context.Background()

	chA, unsubA, err := b.Subscribe(ctx, "room-a")
	if err != nil {
		t.Fatalf("Subscribe(room-a) error = %v", err)
	}
	defer unsubA()
	chB, unsubB, err := b.Subscribe(ctx, "room-b")
	if err != nil {
		t.Fatalf("Subscribe(room-b) error = %v", err)
	}
	defer unsubB()

	if err := b.Publish(ctx, model.ChatMessage{RoomID: "room-a", Body: "only for a"}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case got := <-chA:
		if got.Body != "only for a" {
			t.Errorf("room-a got %q, want %q", got.Body, "only for a")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for room-a's message")
	}

	select {
	case got := <-chB:
		t.Errorf("room-b should not receive room-a's message, got %+v", got)
	case <-time.After(100 * time.Millisecond):
		// expected: nothing arrives
	}
}
