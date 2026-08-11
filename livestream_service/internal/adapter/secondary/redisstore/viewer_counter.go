package redisstore

import (
	"context"
	"strconv"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

func viewerKey(roomID string) string { return "viewers:" + roomID }

type viewerCounter struct {
	client *redis.Client
}

func NewViewerCounter(client *redis.Client) port.ViewerCounter {
	return &viewerCounter{client: client}
}

func (c *viewerCounter) Heartbeat(ctx context.Context, roomID, sessionID string, window time.Duration) error {
	key := viewerKey(roomID)
	pipe := c.client.TxPipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(time.Now().UnixMilli()), Member: sessionID})
	pipe.Expire(ctx, key, window*2)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *viewerCounter) Get(ctx context.Context, roomID string, window time.Duration) (int64, error) {
	key := viewerKey(roomID)
	cutoff := time.Now().Add(-window).UnixMilli()

	pipe := c.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(cutoff, 10))
	card := pipe.ZCard(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return card.Val(), nil
}
