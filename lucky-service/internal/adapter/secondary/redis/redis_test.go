package rediscache

import (
	"context"
	"testing"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type countingRepo struct {
	port.CampaignRepository
	running []model.Campaign
	calls   int
}

func (r *countingRepo) Running(context.Context, time.Time) ([]model.Campaign, error) {
	r.calls++
	return r.running, nil
}

func (r *countingRepo) SetStatus(context.Context, int64, model.CampaignStatus) error { return nil }

func (r *countingRepo) Update(context.Context, *model.Campaign) error { return nil }

func (r *countingRepo) Delete(context.Context, int64) error { return nil }

func client(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func TestCampaignCache(t *testing.T) {
	mr, rdb := client(t)
	base := &countingRepo{running: []model.Campaign{{ID: 1, GameType: model.GameWheel, Prizes: []model.Prize{{ID: 9, Name: "x"}}}}}
	c := NewCampaigns(base, rdb, "t:", time.Minute, time.Second)
	ctx := context.Background()
	for range 3 {
		got, err := c.Running(ctx, time.Now())
		if err != nil || len(got) != 1 || got[0].Prizes[0].ID != 9 {
			t.Fatalf("running: %+v %v", got, err)
		}
	}
	if base.calls != 1 {
		t.Fatalf("cache must absorb repeated reads, db calls = %d", base.calls)
	}
	if err := c.SetStatus(ctx, 1, model.StatusPaused); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Running(ctx, time.Now()); err != nil || base.calls != 2 {
		t.Fatalf("status change must invalidate: calls=%d %v", base.calls, err)
	}
	mr.FastForward(2 * time.Minute)
	if _, err := c.Running(ctx, time.Now()); err != nil || base.calls != 3 {
		t.Fatalf("ttl expiry: calls=%d %v", base.calls, err)
	}
	mr.Close()
	got, err := c.Running(ctx, time.Now())
	if err != nil || len(got) != 1 || base.calls != 4 {
		t.Fatalf("redis outage must fall through to the database: %+v %v", got, err)
	}
}

func TestWinners(t *testing.T) {
	_, rdb := client(t)
	w := NewWinners(rdb, "t:", 3, time.Second)
	ctx := context.Background()
	for i := range 5 {
		if err := w.Push(ctx, model.Winner{CampaignID: 7, Points: int64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	_ = w.Push(ctx, model.Winner{CampaignID: 8, Points: 99})
	got, err := w.Recent(ctx, 7, 10)
	if err != nil || len(got) != 3 || got[0].Points != 4 || got[2].Points != 2 {
		t.Fatalf("feed keeps the newest %d: %+v %v", 3, got, err)
	}
	if ttl := rdb.TTL(ctx, "t:winners:7").Val(); ttl <= 0 {
		t.Fatalf("feed must expire, ttl=%v", ttl)
	}
}

func TestRateLimiter(t *testing.T) {
	_, rdb := client(t)
	l := NewRateLimiter(rdb, "t:", time.Second)
	ctx := context.Background()
	for i := 1; i <= 4; i++ {
		ok, err := l.Allow(ctx, "play:1", 3, time.Minute)
		if err != nil || ok != (i <= 3) {
			t.Fatalf("attempt %d: %v %v", i, ok, err)
		}
	}
	if ok, _ := l.Allow(ctx, "play:2", 3, time.Minute); !ok {
		t.Fatal("limits are per key")
	}
}

func TestSignals(t *testing.T) {
	mr, rdb := client(t)
	s := NewSignals(rdb, "t:", time.Second)
	ctx := context.Background()
	for i, u := range []string{"1", "2", "1", "3"} {
		n, err := s.Observe(ctx, model.SignalDevice, "device-x", u)
		if want := []int{1, 2, 2, 3}[i]; err != nil || n != want {
			t.Fatalf("observe %d: %d %v, want %d", i, n, err, want)
		}
	}
	if n, _ := s.Observe(ctx, model.SignalIP, "device-x", "9"); n != 1 {
		t.Fatalf("device and ip signals are separate, got %d", n)
	}
	mr.FastForward(25 * time.Hour)
	if n, _ := s.Observe(ctx, model.SignalIP, "device-x", "10"); n != 1 {
		t.Fatalf("ip window is one day, got %d", n)
	}
	if n, _ := s.Observe(ctx, model.SignalDevice, "device-x", "4"); n != 4 {
		t.Fatalf("device window is thirty days, got %d", n)
	}
	for _, k := range mr.Keys() {
		if len(k) > 64 {
			t.Fatalf("keys hash the raw value: %s", k)
		}
	}
}
