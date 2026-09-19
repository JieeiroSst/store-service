package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
)

type fakeUserClient struct {
	mu        sync.Mutex
	users     map[string]*model.Author
	getCalls  int32
	batchArgs [][]string
}

func newFakeUserClient() *fakeUserClient {
	return &fakeUserClient{users: make(map[string]*model.Author)}
}

func (f *fakeUserClient) GetUser(ctx context.Context, userID string) (*model.Author, error) {
	atomic.AddInt32(&f.getCalls, 1)
	f.mu.Lock()
	defer f.mu.Unlock()
	author, ok := f.users[userID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return author, nil
}

func (f *fakeUserClient) GetUsers(ctx context.Context, userIDs []string) (map[string]*model.Author, error) {
	f.mu.Lock()
	f.batchArgs = append(f.batchArgs, userIDs)
	f.mu.Unlock()

	result := make(map[string]*model.Author, len(userIDs))
	for _, id := range userIDs {
		if a, ok := f.users[id]; ok {
			result[id] = a
		}
	}
	return result, nil
}

func TestCachedUserClientCachesGetUserOnMiss(t *testing.T) {
	next := newFakeUserClient()
	next.users["user-1"] = &model.Author{ID: "user-1", Username: "alice"}
	client := NewCachedUserClient(next, newTestRedisClient(t))

	if _, err := client.GetUser(context.Background(), "user-1"); err != nil {
		t.Fatalf("first GetUser() error = %v", err)
	}
	if _, err := client.GetUser(context.Background(), "user-1"); err != nil {
		t.Fatalf("second GetUser() error = %v", err)
	}

	if got := atomic.LoadInt32(&next.getCalls); got != 1 {
		t.Errorf("underlying GetUser calls = %d, want 1", got)
	}
}

func TestCachedUserClientGetUsersOnlyFetchesMissingIDs(t *testing.T) {
	next := newFakeUserClient()
	next.users["user-1"] = &model.Author{ID: "user-1", Username: "alice"}
	next.users["user-2"] = &model.Author{ID: "user-2", Username: "bob"}
	client := NewCachedUserClient(next, newTestRedisClient(t))
	ctx := context.Background()

	// Warm the cache for user-1 only.
	if _, err := client.GetUser(ctx, "user-1"); err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}

	result, err := client.GetUsers(ctx, []string{"user-1", "user-2"})
	if err != nil {
		t.Fatalf("GetUsers() error = %v", err)
	}
	if result["user-1"] == nil || result["user-1"].Username != "alice" {
		t.Errorf("result[user-1] = %+v, want alice", result["user-1"])
	}
	if result["user-2"] == nil || result["user-2"].Username != "bob" {
		t.Errorf("result[user-2] = %+v, want bob", result["user-2"])
	}

	next.mu.Lock()
	defer next.mu.Unlock()
	if len(next.batchArgs) != 1 || len(next.batchArgs[0]) != 1 || next.batchArgs[0][0] != "user-2" {
		t.Errorf("underlying GetUsers batch args = %+v, want a single call for just [user-2]", next.batchArgs)
	}
}
