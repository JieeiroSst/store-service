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
	userCacheTTL    = 5 * time.Minute
	userCacheJitter = 60 // seconds
)

func userCacheKey(id string) string { return "cache:user:" + id }

type cachedUserClient struct {
	next  port.UserClient
	redis *redis.Client
	group singleflight.Group
}

func NewCachedUserClient(next port.UserClient, redisClient *redis.Client) port.UserClient {
	return &cachedUserClient{next: next, redis: redisClient}
}

func (c *cachedUserClient) GetUser(ctx context.Context, userID string) (*model.Author, error) {
	key := userCacheKey(userID)
	if cached, err := c.redis.Get(ctx, key).Result(); err == nil {
		var author model.Author
		if jsonErr := json.Unmarshal([]byte(cached), &author); jsonErr == nil {
			return &author, nil
		}
	}

	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		author, err := c.next.GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		c.set(ctx, key, author)
		return author, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*model.Author), nil
}

func (c *cachedUserClient) set(ctx context.Context, key string, author *model.Author) {
	data, err := json.Marshal(author)
	if err != nil {
		return
	}
	jitter := time.Duration(rand.Intn(userCacheJitter)) * time.Second
	c.redis.Set(ctx, key, data, userCacheTTL+jitter)
}

func (c *cachedUserClient) GetUsers(ctx context.Context, userIDs []string) (map[string]*model.Author, error) {
	result := make(map[string]*model.Author, len(userIDs))
	var missing []string

	for _, id := range userIDs {
		if cached, err := c.redis.Get(ctx, userCacheKey(id)).Result(); err == nil {
			var author model.Author
			if json.Unmarshal([]byte(cached), &author) == nil {
				result[id] = &author
				continue
			}
		}
		missing = append(missing, id)
	}
	if len(missing) == 0 {
		return result, nil
	}

	fetched, err := c.next.GetUsers(ctx, missing)
	if err != nil {
		return result, nil
	}
	for id, author := range fetched {
		result[id] = author
		c.set(ctx, userCacheKey(id), author)
	}
	return result, nil
}
