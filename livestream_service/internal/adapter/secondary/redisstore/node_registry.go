package redisstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

const capacityZSetKey = "nodes:capacity"

func nodeKey(nodeID string) string          { return "node:" + nodeID }
func assignmentKey(streamKey string) string { return "stream:" + streamKey + ":node" }

type nodeRegistry struct {
	client *redis.Client
}

func NewNodeRegistry(client *redis.Client) port.NodeRegistry {
	return &nodeRegistry{client: client}
}

func (r *nodeRegistry) Heartbeat(ctx context.Context, node model.TranscodeNode, ttl time.Duration) error {
	key := nodeKey(node.ID)
	pipe := r.client.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"id":          node.ID,
		"addr":        node.Addr,
		"httpAddr":    node.HTTPAddr,
		"active":      node.ActiveStreams,
		"max":         node.MaxStreams,
		"capacity":    node.Capacity,
		"loadPerCore": node.LoadPerCore,
		"ts":          time.Now().Unix(),
	})
	pipe.Expire(ctx, key, ttl)
	pipe.ZAdd(ctx, capacityZSetKey, redis.Z{Score: float64(node.Capacity), Member: node.ID})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *nodeRegistry) TopCandidates(ctx context.Context, n int) ([]string, error) {
	return r.client.ZRevRange(ctx, capacityZSetKey, 0, int64(n-1)).Result()
}

func (r *nodeRegistry) GetNode(ctx context.Context, nodeID string) (*model.TranscodeNode, bool, error) {
	fields, err := r.client.HGetAll(ctx, nodeKey(nodeID)).Result()
	if err != nil {
		return nil, false, err
	}
	if len(fields) == 0 {
		return nil, false, nil
	}

	node := &model.TranscodeNode{
		ID:            fields["id"],
		Addr:          fields["addr"],
		HTTPAddr:      fields["httpAddr"],
		ActiveStreams: atoiDefault(fields["active"], 0),
		MaxStreams:    atoiDefault(fields["max"], 0),
		Capacity:      atoiDefault(fields["capacity"], 0),
		LoadPerCore:   atofDefault(fields["loadPerCore"], 0),
	}
	if ts, err := strconv.ParseInt(fields["ts"], 10, 64); err == nil {
		node.LastHeartbeat = time.Unix(ts, 0)
	}
	return node, true, nil
}

func (r *nodeRegistry) ReserveCapacity(ctx context.Context, nodeID string) (bool, error) {
	newScore, err := r.client.ZIncrBy(ctx, capacityZSetKey, -1, nodeID).Result()
	if err != nil {
		return false, err
	}
	if newScore < 0 {
		if _, revertErr := r.client.ZIncrBy(ctx, capacityZSetKey, 1, nodeID).Result(); revertErr != nil {
			return false, fmt.Errorf("revert over-reserved capacity: %w", revertErr)
		}
		return false, nil
	}
	if err := r.client.HIncrBy(ctx, nodeKey(nodeID), "capacity", -1).Err(); err != nil {
		return false, err
	}
	return true, nil
}

func (r *nodeRegistry) ReleaseCapacity(ctx context.Context, nodeID string) error {
	if err := r.client.ZIncrBy(ctx, capacityZSetKey, 1, nodeID).Err(); err != nil {
		return err
	}
	return r.client.HIncrBy(ctx, nodeKey(nodeID), "capacity", 1).Err()
}

func (r *nodeRegistry) GetAssignment(ctx context.Context, streamKey string) (string, bool, error) {
	v, err := r.client.Get(ctx, assignmentKey(streamKey)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (r *nodeRegistry) SetAssignment(ctx context.Context, streamKey, nodeID string, ttl time.Duration) error {
	return r.client.Set(ctx, assignmentKey(streamKey), nodeID, ttl).Err()
}

func (r *nodeRegistry) ClearAssignment(ctx context.Context, streamKey string) error {
	return r.client.Del(ctx, assignmentKey(streamKey)).Err()
}

func atoiDefault(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return fallback
}

func atofDefault(s string, fallback float64) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return fallback
}
