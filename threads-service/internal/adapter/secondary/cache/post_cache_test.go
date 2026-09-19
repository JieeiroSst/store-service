package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

// fakePostRepository is a minimal PostRepository double that counts
// GetByID calls, so tests can assert how many times Postgres was actually
// hit rather than just checking the returned value.
type fakePostRepository struct {
	mu       sync.Mutex
	posts    map[string]*model.Post
	getCalls int32
	getDelay time.Duration
}

func newFakePostRepository() *fakePostRepository {
	return &fakePostRepository{posts: make(map[string]*model.Post)}
}

func (f *fakePostRepository) Create(ctx context.Context, post *model.Post) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.posts[post.ID] = post
	return nil
}

func (f *fakePostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	atomic.AddInt32(&f.getCalls, 1)
	if f.getDelay > 0 {
		time.Sleep(f.getDelay)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	post, ok := f.posts[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	got := *post
	return &got, nil
}

func (f *fakePostRepository) getCallCount() int32 { return atomic.LoadInt32(&f.getCalls) }

func (f *fakePostRepository) ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error) {
	return nil, "", nil
}
func (f *fakePostRepository) ListByAuthors(ctx context.Context, authorIDs []string, cursor string, limit int) ([]model.Post, string, error) {
	return nil, "", nil
}
func (f *fakePostRepository) Delete(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.posts, id)
	return nil
}
func (f *fakePostRepository) IncrementLikeCount(ctx context.Context, id string, delta int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if post, ok := f.posts[id]; ok {
		post.LikeCount += delta
	}
	return nil
}
func (f *fakePostRepository) IncrementCommentCount(ctx context.Context, id string, delta int) error {
	return nil
}
func (f *fakePostRepository) IncrementRepostCount(ctx context.Context, id string, delta int) error {
	return nil
}
func (f *fakePostRepository) FindRepostBy(ctx context.Context, userID, originalPostID string) (*model.Post, error) {
	return nil, port.ErrNotFound
}

func TestCachedPostRepositoryCachesOnMiss(t *testing.T) {
	next := newFakePostRepository()
	next.posts["post-1"] = &model.Post{ID: "post-1", Content: "hello"}
	repo := NewCachedPostRepository(next, newTestRedisClient(t))

	if _, err := repo.GetByID(context.Background(), "post-1"); err != nil {
		t.Fatalf("first GetByID() error = %v", err)
	}
	if _, err := repo.GetByID(context.Background(), "post-1"); err != nil {
		t.Fatalf("second GetByID() error = %v", err)
	}

	if got := next.getCallCount(); got != 1 {
		t.Errorf("underlying GetByID calls = %d, want 1 (second read should be a cache hit)", got)
	}
}

func TestCachedPostRepositoryInvalidatesOnWrite(t *testing.T) {
	next := newFakePostRepository()
	next.posts["post-1"] = &model.Post{ID: "post-1", LikeCount: 0}
	repo := NewCachedPostRepository(next, newTestRedisClient(t))
	ctx := context.Background()

	post, err := repo.GetByID(ctx, "post-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if post.LikeCount != 0 {
		t.Fatalf("LikeCount = %d, want 0", post.LikeCount)
	}

	if err := repo.IncrementLikeCount(ctx, "post-1", 1); err != nil {
		t.Fatalf("IncrementLikeCount() error = %v", err)
	}

	post, err = repo.GetByID(ctx, "post-1")
	if err != nil {
		t.Fatalf("GetByID() after write error = %v", err)
	}
	if post.LikeCount != 1 {
		t.Errorf("LikeCount = %d after IncrementLikeCount, want 1 (stale cache was not invalidated)", post.LikeCount)
	}
	if got := next.getCallCount(); got != 2 {
		t.Errorf("underlying GetByID calls = %d, want 2 (cache must be bypassed after invalidation)", got)
	}
}

func TestCachedPostRepositoryCollapsesConcurrentMissesViaSingleflight(t *testing.T) {
	next := newFakePostRepository()
	next.getDelay = 50 * time.Millisecond
	next.posts["post-1"] = &model.Post{ID: "post-1"}
	repo := NewCachedPostRepository(next, newTestRedisClient(t))

	const concurrency = 20
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			if _, err := repo.GetByID(context.Background(), "post-1"); err != nil {
				t.Errorf("GetByID() error = %v", err)
			}
		}()
	}
	wg.Wait()

	// Without singleflight, 20 concurrent cache misses for the same key
	// would each fall through to the DB. With it, they collapse into one.
	if got := next.getCallCount(); got != 1 {
		t.Errorf("underlying GetByID calls = %d, want 1 for %d concurrent readers of the same cold key", got, concurrency)
	}
}
