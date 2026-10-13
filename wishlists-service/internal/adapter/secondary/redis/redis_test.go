package rediscache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/pkg/credential"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

var (
	ctx  = credential.With(context.Background(), credential.Credential{Bearer: "access"})
	anon = context.Background()
)

func setup(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func seed(t *testing.T, base *memory.Repository, vis model.Visibility) *model.Wishlist {
	w, _ := model.NewWishlist("1", "List", vis)
	if err := base.Create(ctx, w); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestRepositoryCachesAndInvalidates(t *testing.T) {
	_, rdb := setup(t)
	base := memory.New()
	repo := NewRepository(base, rdb, "t:", time.Minute, time.Minute, time.Second)
	w := seed(t, base, model.VisibilityShared)

	if _, err := repo.Get(ctx, w.ID); err != nil {
		t.Fatal(err)
	}
	_, _ = base.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 1, Title: "x", Quantity: 1})
	cached, _ := repo.Get(ctx, w.ID)
	if len(cached.Lines) != 0 {
		t.Fatalf("expected cached copy without the line written behind the cache, got %+v", cached.Lines)
	}

	if _, err := repo.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 2, Title: "y", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	fresh, _ := repo.Get(ctx, w.ID)
	if len(fresh.Lines) != 2 {
		t.Fatalf("write through the decorator must invalidate, got %+v", fresh.Lines)
	}

	byKey, err := repo.GetByKey(ctx, w.Key)
	if err != nil || byKey.ID != w.ID {
		t.Fatalf("get by key: %+v %v", byKey, err)
	}
	oldKey := w.Key
	_ = w.RegenerateKey()
	if err := repo.Update(ctx, w); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, oldKey); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("old share key must stop working after regeneration: %v", err)
	}
	if got, err := repo.GetByKey(ctx, w.Key); err != nil || got.ID != w.ID {
		t.Fatalf("new key: %+v %v", got, err)
	}

	if err := repo.Delete(ctx, w.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, w.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("deleted wishlist served from cache: %v", err)
	}
}

func TestStaleReadCannotOverwriteNewerData(t *testing.T) {
	_, rdb := setup(t)
	base := memory.New()
	repo := NewRepository(base, rdb, "t:", time.Minute, time.Minute, time.Second)
	w := seed(t, base, model.VisibilityPrivate)

	gen, _ := repo.load(ctx, w.ID)
	stale, _ := base.Get(ctx, w.ID)
	if _, err := repo.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 1, Title: "x", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	repo.store(ctx, stale, gen)

	got, _ := repo.Get(ctx, w.ID)
	if len(got.Lines) != 1 {
		t.Fatalf("a read that started before the write must not repopulate the cache, got %+v", got.Lines)
	}
}

func TestMoveInvalidatesBothWishlists(t *testing.T) {
	_, rdb := setup(t)
	base := memory.New()
	repo := NewRepository(base, rdb, "t:", time.Minute, time.Minute, time.Second)
	a := seed(t, base, model.VisibilityPrivate)
	b := seed(t, base, model.VisibilityPrivate)
	l, _ := repo.AddLine(ctx, &model.Line{WishlistID: a.ID, ProductID: 1, Title: "x", Quantity: 1})
	_, _ = repo.Get(ctx, a.ID)
	_, _ = repo.Get(ctx, b.ID)
	if _, err := repo.MoveLine(ctx, a.ID, l.ID, b.ID); err != nil {
		t.Fatal(err)
	}
	ga, _ := repo.Get(ctx, a.ID)
	gb, _ := repo.Get(ctx, b.ID)
	if len(ga.Lines) != 0 || len(gb.Lines) != 1 {
		t.Fatalf("after move: a=%+v b=%+v", ga.Lines, gb.Lines)
	}
	if n, _ := repo.ClearLines(ctx, b.ID); n != 1 {
		t.Fatalf("clear: %d", n)
	}
	if gb, _ := repo.Get(ctx, b.ID); len(gb.Lines) != 0 {
		t.Fatalf("clear must invalidate: %+v", gb.Lines)
	}
}

func TestPopularProductsCached(t *testing.T) {
	mr, rdb := setup(t)
	base := memory.New()
	repo := NewRepository(base, rdb, "t:", time.Minute, time.Minute, time.Second)
	w := seed(t, base, model.VisibilityPrivate)
	_, _ = base.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 1, Title: "x", Quantity: 1})
	first, _ := repo.PopularProducts(ctx, 5)
	_, _ = base.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 2, Title: "y", Quantity: 1})
	second, _ := repo.PopularProducts(ctx, 5)
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected cached stats: %+v %+v", first, second)
	}
	mr.FastForward(2 * time.Minute)
	third, _ := repo.PopularProducts(ctx, 5)
	if len(third) != 2 {
		t.Fatalf("expected refreshed stats after ttl: %+v", third)
	}
}

func TestFailOpenWhenRedisIsDown(t *testing.T) {
	mr, rdb := setup(t)
	base := memory.New()
	repo := NewRepository(base, rdb, "t:", time.Minute, time.Minute, 100*time.Millisecond)
	users := NewUsers(memory.NewUsers(model.Owner{ID: "1", Username: "alice"}), rdb, "t:", time.Minute, time.Minute, 100*time.Millisecond)
	limiter := NewRateLimiter(rdb, "t:", 100*time.Millisecond)
	w := seed(t, base, model.VisibilityPrivate)
	mr.Close()

	if _, err := repo.Get(ctx, w.ID); err != nil {
		t.Fatalf("get with redis down: %v", err)
	}
	if _, err := repo.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 1, Title: "x", Quantity: 1}); err != nil {
		t.Fatalf("write with redis down: %v", err)
	}
	if _, err := repo.GetByKey(ctx, w.Key); err != nil {
		t.Fatalf("get by key with redis down: %v", err)
	}
	if u, err := users.GetUser(ctx, "1"); err != nil || u.Username != "alice" {
		t.Fatalf("user lookup with redis down: %+v %v", u, err)
	}
	if ok, err := limiter.Allow(ctx, "u", 1, time.Minute); !ok || err == nil {
		t.Fatalf("limiter should allow and report the error: %v %v", ok, err)
	}
}

func TestUsersCache(t *testing.T) {
	mr, rdb := setup(t)
	base := memory.NewUsers(model.Owner{ID: "1", Username: "alice", Active: true})
	users := NewUsers(base, rdb, "t:", time.Minute, 10*time.Second, time.Second)

	for i := 0; i < 3; i++ {
		u, err := users.GetUser(ctx, "1")
		if err != nil || u.Username != "alice" {
			t.Fatalf("get user: %+v %v", u, err)
		}
	}
	for i := 0; i < 3; i++ {
		if _, err := users.GetUser(ctx, "404"); !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("missing user: %v", err)
		}
	}
	if base.Calls != 2 {
		t.Fatalf("expected 2 user-service calls (hit + negative cache), got %d", base.Calls)
	}
	if u, err := users.GetUser(anon, "1"); err != nil || u.Username != "alice" {
		t.Fatalf("anonymous readers may use cached public profiles: %+v %v", u, err)
	}
	if _, err := users.GetUser(anon, "2"); !errors.Is(err, model.ErrUnauthenticated) || base.Calls != 2 {
		t.Fatalf("anonymous cache miss must not call user-service: %v calls=%d", err, base.Calls)
	}
	mr.FastForward(11 * time.Second)
	_, _ = users.GetUser(ctx, "404")
	if base.Calls != 3 {
		t.Fatalf("negative entry should expire, calls=%d", base.Calls)
	}

	base.Err = model.ErrUnavailable
	if _, err := users.GetUser(ctx, "7"); !errors.Is(err, model.ErrUnavailable) {
		t.Fatalf("unavailable: %v", err)
	}
	base.Err = nil
	if _, err := users.GetUser(ctx, "7"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("errors must not be cached: %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	_, rdb := setup(t)
	l := NewRateLimiter(rdb, "t:", time.Second)
	now := time.Unix(1_700_000_000, 0)
	l.now = func() time.Time { return now }
	for i := 0; i < 2; i++ {
		if ok, err := l.Allow(ctx, "u1", 2, time.Minute); !ok || err != nil {
			t.Fatalf("request %d: %v %v", i, ok, err)
		}
	}
	if ok, _ := l.Allow(ctx, "u1", 2, time.Minute); ok {
		t.Fatal("third request should be limited")
	}
	if ok, _ := l.Allow(ctx, "u2", 2, time.Minute); !ok {
		t.Fatal("limits are per key")
	}
	now = now.Add(time.Minute)
	if ok, _ := l.Allow(ctx, "u1", 2, time.Minute); !ok {
		t.Fatal("new window should reset the limit")
	}
	if ok, _ := l.Allow(ctx, "u1", 0, time.Minute); !ok {
		t.Fatal("limit 0 disables limiting")
	}
}
