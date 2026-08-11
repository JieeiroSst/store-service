package redisstore

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	_, client := newTestRedisServer(t)
	return client
}

// newTestRedisServer also returns the underlying miniredis instance, for
// tests that need to advance its virtual clock (miniredis doesn't expire
// TTL'd keys just because real time passes - see mr.FastForward).
func newTestRedisServer(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

func TestViewerCounterCountsDistinctSessions(t *testing.T) {
	vc := NewViewerCounter(newTestRedisClient(t))
	ctx := context.Background()
	window := time.Minute

	for _, session := range []string{"session-a", "session-b", "session-c"} {
		if err := vc.Heartbeat(ctx, "room-1", session, window); err != nil {
			t.Fatalf("Heartbeat(%s) error = %v", session, err)
		}
	}

	count, err := vc.Get(ctx, "room-1", window)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Get() = %d, want 3 distinct sessions", count)
	}
}

func TestViewerCounterRepeatedHeartbeatDoesNotDoubleCount(t *testing.T) {
	vc := NewViewerCounter(newTestRedisClient(t))
	ctx := context.Background()
	window := time.Minute

	for i := 0; i < 5; i++ {
		if err := vc.Heartbeat(ctx, "room-1", "session-a", window); err != nil {
			t.Fatalf("Heartbeat() error = %v", err)
		}
	}

	count, err := vc.Get(ctx, "room-1", window)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Get() = %d, want 1 (repeated heartbeats from the same session)", count)
	}
}

func TestViewerCounterAgesOutStaleSessions(t *testing.T) {
	vc := NewViewerCounter(newTestRedisClient(t))
	ctx := context.Background()
	window := 50 * time.Millisecond

	if err := vc.Heartbeat(ctx, "room-1", "session-a", window); err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}
	if count, err := vc.Get(ctx, "room-1", window); err != nil || count != 1 {
		t.Fatalf("Get() = %d, err = %v; want 1 immediately after heartbeat", count, err)
	}

	time.Sleep(3 * window)

	count, err := vc.Get(ctx, "room-1", window)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Get() = %d, want 0 - session should have aged out of the window", count)
	}
}

func TestViewerCounterGetOnUnknownRoomIsZero(t *testing.T) {
	vc := NewViewerCounter(newTestRedisClient(t))

	count, err := vc.Get(context.Background(), "never-heard-of-it", time.Minute)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Get() = %d, want 0 for a room with no heartbeats", count)
	}
}
