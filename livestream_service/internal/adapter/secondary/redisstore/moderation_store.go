package redisstore

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

func banKey(roomID, userID string) string { return "chatban:" + roomID + ":" + userID }

// moderationStore is a plain TTL'd key per ban - it lapses on its own, no
// cleanup job needed. A permanent ban is just a very long TTL.
type moderationStore struct {
	client *redis.Client
}

func NewModerationStore(client *redis.Client) port.ModerationStore {
	return &moderationStore{client: client}
}

func (m *moderationStore) Ban(ctx context.Context, roomID, userID string, ttl time.Duration) error {
	return m.client.Set(ctx, banKey(roomID, userID), "1", ttl).Err()
}

func (m *moderationStore) Unban(ctx context.Context, roomID, userID string) error {
	return m.client.Del(ctx, banKey(roomID, userID)).Err()
}

func (m *moderationStore) IsBanned(ctx context.Context, roomID, userID string) (bool, error) {
	_, err := m.client.Get(ctx, banKey(roomID, userID)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
