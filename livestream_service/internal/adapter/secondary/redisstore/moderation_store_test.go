package redisstore

import (
	"context"
	"testing"
	"time"
)

func TestModerationStoreBanThenIsBanned(t *testing.T) {
	m := NewModerationStore(newTestRedisClient(t))
	ctx := context.Background()

	if banned, err := m.IsBanned(ctx, "room-1", "troll"); err != nil || banned {
		t.Fatalf("IsBanned() = %v, %v before any ban; want false, nil", banned, err)
	}

	if err := m.Ban(ctx, "room-1", "troll", time.Minute); err != nil {
		t.Fatalf("Ban() error = %v", err)
	}
	if banned, err := m.IsBanned(ctx, "room-1", "troll"); err != nil || !banned {
		t.Fatalf("IsBanned() = %v, %v after Ban(); want true, nil", banned, err)
	}
}

func TestModerationStoreUnban(t *testing.T) {
	m := NewModerationStore(newTestRedisClient(t))
	ctx := context.Background()

	_ = m.Ban(ctx, "room-1", "troll", time.Minute)
	if err := m.Unban(ctx, "room-1", "troll"); err != nil {
		t.Fatalf("Unban() error = %v", err)
	}
	if banned, _ := m.IsBanned(ctx, "room-1", "troll"); banned {
		t.Error("expected troll to no longer be banned after Unban")
	}
}

func TestModerationStoreBanExpiresOnItsOwn(t *testing.T) {
	mr, client := newTestRedisServer(t)
	m := NewModerationStore(client)
	ctx := context.Background()

	ttl := time.Minute
	if err := m.Ban(ctx, "room-1", "troll", ttl); err != nil {
		t.Fatalf("Ban() error = %v", err)
	}
	// The key's TTL is a real Redis expiry, not the application-computed
	// sliding window ViewerCounter uses - miniredis only ages it out when
	// its virtual clock is advanced, not when real time passes.
	mr.FastForward(ttl + time.Second)

	if banned, err := m.IsBanned(ctx, "room-1", "troll"); err != nil || banned {
		t.Fatalf("IsBanned() = %v, %v after TTL expiry; want false, nil", banned, err)
	}
}

func TestModerationStoreBansAreScopedPerRoom(t *testing.T) {
	m := NewModerationStore(newTestRedisClient(t))
	ctx := context.Background()

	_ = m.Ban(ctx, "room-1", "troll", time.Minute)
	if banned, _ := m.IsBanned(ctx, "room-2", "troll"); banned {
		t.Error("expected a ban in room-1 to not apply to room-2")
	}
}
