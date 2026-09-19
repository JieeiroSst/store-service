package trending

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestTopTagsRanksByUsageCount(t *testing.T) {
	s := NewTrendingTagsStore(newTestRedisClient(t))
	ctx := context.Background()

	if err := s.IncrementTags(ctx, []string{"go", "redis"}); err != nil {
		t.Fatalf("IncrementTags() error = %v", err)
	}
	if err := s.IncrementTags(ctx, []string{"go"}); err != nil {
		t.Fatalf("IncrementTags() error = %v", err)
	}

	top, err := s.TopTags(ctx, 10)
	if err != nil {
		t.Fatalf("TopTags() error = %v", err)
	}
	if len(top) != 2 || top[0].Name != "go" || top[0].Count != 2 {
		t.Fatalf("TopTags() = %+v, want go(2) ranked first", top)
	}
	if top[1].Name != "redis" || top[1].Count != 1 {
		t.Fatalf("TopTags()[1] = %+v, want redis(1)", top[1])
	}
}

func TestTopTagsRespectsLimit(t *testing.T) {
	s := NewTrendingTagsStore(newTestRedisClient(t))
	ctx := context.Background()

	_ = s.IncrementTags(ctx, []string{"a", "b", "c"})

	top, err := s.TopTags(ctx, 2)
	if err != nil {
		t.Fatalf("TopTags() error = %v", err)
	}
	if len(top) != 2 {
		t.Fatalf("TopTags() returned %d entries, want 2", len(top))
	}
}

func TestIncrementTagsIsANoOpForEmptyInput(t *testing.T) {
	s := NewTrendingTagsStore(newTestRedisClient(t))
	if err := s.IncrementTags(context.Background(), nil); err != nil {
		t.Fatalf("IncrementTags(nil) error = %v", err)
	}
}
