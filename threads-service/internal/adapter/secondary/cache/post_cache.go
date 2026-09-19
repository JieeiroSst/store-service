package cache

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	postCacheTTL    = 60 * time.Second
	postCacheJitter = 10 // seconds
)

func postCacheKey(id string) string { return "cache:post:" + id }

type cachedPostRepository struct {
	next  port.PostRepository
	redis *redis.Client
	group singleflight.Group
}

func NewCachedPostRepository(next port.PostRepository, redisClient *redis.Client) port.PostRepository {
	return &cachedPostRepository{next: next, redis: redisClient}
}

func (c *cachedPostRepository) Create(ctx context.Context, post *model.Post) error {
	return c.next.Create(ctx, post)
}

func (c *cachedPostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	key := postCacheKey(id)
	if cached, err := c.redis.Get(ctx, key).Result(); err == nil {
		var post model.Post
		if jsonErr := json.Unmarshal([]byte(cached), &post); jsonErr == nil {
			return &post, nil
		}
	}

	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		post, err := c.next.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		c.set(ctx, key, post)
		return post, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*model.Post), nil
}

func (c *cachedPostRepository) set(ctx context.Context, key string, post *model.Post) {
	data, err := json.Marshal(post)
	if err != nil {
		return
	}
	jitter := time.Duration(rand.Intn(postCacheJitter)) * time.Second
	c.redis.Set(ctx, key, data, postCacheTTL+jitter)
}

func (c *cachedPostRepository) invalidate(ctx context.Context, id string) {
	key := postCacheKey(id)
	c.redis.Del(ctx, key)
	// Delayed double-delete - see package doc, point 2.
	go func() {
		time.Sleep(500 * time.Millisecond)
		c.redis.Del(context.Background(), key)
	}()
}

func (c *cachedPostRepository) Delete(ctx context.Context, id string) error {
	if err := c.next.Delete(ctx, id); err != nil {
		return err
	}
	c.invalidate(ctx, id)
	return nil
}

func (c *cachedPostRepository) IncrementLikeCount(ctx context.Context, id string, delta int) error {
	if err := c.next.IncrementLikeCount(ctx, id, delta); err != nil {
		return err
	}
	c.invalidate(ctx, id)
	return nil
}

func (c *cachedPostRepository) IncrementCommentCount(ctx context.Context, id string, delta int) error {
	if err := c.next.IncrementCommentCount(ctx, id, delta); err != nil {
		return err
	}
	c.invalidate(ctx, id)
	return nil
}

func (c *cachedPostRepository) IncrementRepostCount(ctx context.Context, id string, delta int) error {
	if err := c.next.IncrementRepostCount(ctx, id, delta); err != nil {
		return err
	}
	c.invalidate(ctx, id)
	return nil
}

func (c *cachedPostRepository) FindRepostBy(ctx context.Context, userID, originalPostID string) (*model.Post, error) {
	return c.next.FindRepostBy(ctx, userID, originalPostID)
}

func (c *cachedPostRepository) ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error) {
	return c.next.ListFeed(ctx, authorID, cursor, limit)
}

func (c *cachedPostRepository) ListByAuthors(ctx context.Context, authorIDs []string, cursor string, limit int) ([]model.Post, string, error) {
	return c.next.ListByAuthors(ctx, authorIDs, cursor, limit)
}
