package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestMemoryExpiryAndEviction(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(0, 0)
	m := NewMemory(2)
	m.now = func() time.Time { return now }

	_ = m.Set(ctx, "a", []byte("1"), time.Minute)
	_ = m.Set(ctx, "b", []byte("2"), time.Minute)
	if _, ok, _ := m.Get(ctx, "a"); !ok {
		t.Fatal("a missing")
	}
	_ = m.Set(ctx, "c", []byte("3"), time.Minute)
	if _, ok, _ := m.Get(ctx, "b"); ok {
		t.Fatal("b should have been evicted")
	}
	now = now.Add(time.Minute)
	if _, ok, _ := m.Get(ctx, "a"); ok {
		t.Fatal("a should have expired")
	}
	if m.ll.Len() != 1 {
		t.Fatalf("expired entry not removed, len=%d", m.ll.Len())
	}
}

func TestRedis(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	r := NewRedis(redis.NewClient(&redis.Options{Addr: mr.Addr()}), "serpapi:")
	if _, ok, err := r.Get(ctx, "k"); ok || err != nil {
		t.Fatalf("miss: ok=%v err=%v", ok, err)
	}
	if err := r.Set(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if v, ok, err := r.Get(ctx, "k"); !ok || err != nil || string(v) != "v" {
		t.Fatalf("hit: %q %v %v", v, ok, err)
	}
	if !mr.Exists("serpapi:k") {
		t.Fatal("prefix not applied")
	}
	mr.FastForward(2 * time.Minute)
	if _, ok, _ := r.Get(ctx, "k"); ok {
		t.Fatal("ttl not applied")
	}
}
