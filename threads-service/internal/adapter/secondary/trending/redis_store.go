package trending

import (
	"context"
	"time"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

const bucketTTL = 48 * time.Hour

type redisTrendingStore struct {
	redis *redis.Client
}

func NewTrendingTagsStore(redisClient *redis.Client) port.TrendingTagsStore {
	return &redisTrendingStore{redis: redisClient}
}

func dailyKey(t time.Time) string {
	return "trending:tags:" + t.UTC().Format("20060102")
}

func (s *redisTrendingStore) IncrementTags(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	key := dailyKey(time.Now())
	pipe := s.redis.Pipeline()
	for _, name := range names {
		pipe.ZIncrBy(ctx, key, 1, name)
	}
	pipe.Expire(ctx, key, bucketTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *redisTrendingStore) TopTags(ctx context.Context, limit int) ([]model.TagCount, error) {
	key := dailyKey(time.Now())
	results, err := s.redis.ZRevRangeWithScores(ctx, key, 0, int64(limit)-1).Result()
	if err != nil {
		return nil, err
	}
	tags := make([]model.TagCount, len(results))
	for i, z := range results {
		tags[i] = model.TagCount{Name: z.Member.(string), Count: int64(z.Score)}
	}
	return tags, nil
}

var Module = fx.Options(
	fx.Provide(NewTrendingTagsStore),
)
