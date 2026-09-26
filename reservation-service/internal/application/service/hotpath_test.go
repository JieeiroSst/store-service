package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

func TestSoldOutCache(t *testing.T) {
	now := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	c := NewSoldOutCache(2 * time.Second)
	c.now = func() time.Time { return now }
	s, e := day(12), day(14)

	if c.SoldOut(1, s, e, 1) {
		t.Fatal("unknown range is not sold out")
	}
	c.MarkSoldOut(1, s, e, 2) // 2 rooms could not be had, so 2 or more are sold out; 1 might still be free
	if c.SoldOut(1, s, e, 1) || !c.SoldOut(1, s, e, 2) || !c.SoldOut(1, s, e, 5) {
		t.Fatal("threshold")
	}
	c.MarkSoldOut(1, s, e, 1) // now known for 1 room: covers everything
	if !c.SoldOut(1, s, e, 1) || !c.SoldOut(1, s, e, 3) {
		t.Fatal("smaller threshold must win")
	}
	if c.SoldOut(2, s, e, 1) || c.SoldOut(1, s, day(15), 1) || c.SoldOut(1, day(13), e, 1) {
		t.Fatal("other room types and ranges are unaffected")
	}
	// Rooms coming back invalidate the room type at once.
	c.Release(1)
	if c.SoldOut(1, s, e, 1) {
		t.Fatal("release must clear it")
	}
	// The TTL bounds how long a stale answer lives.
	c.MarkSoldOut(1, s, e, 1)
	now = now.Add(3 * time.Second)
	if c.SoldOut(1, s, e, 1) {
		t.Fatal("expired")
	}
	// A disabled or nil cache never says sold out.
	var nilCache *SoldOutCache
	nilCache.MarkSoldOut(1, s, e, 1)
	if nilCache.SoldOut(1, s, e, 1) || NewSoldOutCache(0).SoldOut(1, s, e, 1) {
		t.Fatal("disabled")
	}
}

func TestSoldOutCacheIsBounded(t *testing.T) {
	c := NewSoldOutCache(time.Minute)
	c.max = 100
	for i := 0; i < 1000; i++ {
		c.MarkSoldOut(1, day(1).AddDate(0, 0, i), day(1).AddDate(0, 0, i+1), 1)
	}
	if len(c.entries) > 100 {
		t.Fatalf("grew to %d", len(c.entries))
	}
}

func TestBulkheadLimitsConcurrencyAndShedsTheRest(t *testing.T) {
	b := NewBulkhead(3, 20*time.Millisecond)
	var inside, peak, shed atomic.Int32
	hold := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			leave, err := b.Enter(context.Background())
			if errors.Is(err, domain.ErrBusy) {
				shed.Add(1)
				return
			}
			n := inside.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			<-hold
			inside.Add(-1)
			leave()
		}()
	}
	time.Sleep(200 * time.Millisecond) // everyone has tried once and the waiting ones timed out
	close(hold)
	wg.Wait()
	if peak.Load() != 3 || shed.Load() != 47 {
		t.Fatalf("peak %d (want 3), shed %d (want 47)", peak.Load(), shed.Load())
	}
	// Slots are given back: a later caller gets in.
	leave, err := b.Enter(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	leave()
	var none *Bulkhead
	if leave, err := none.Enter(context.Background()); err != nil {
		t.Fatal(err)
	} else {
		leave()
	}
}

func TestUserLimiter(t *testing.T) {
	now := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	l := NewUserLimiter(2, 3) // 3 at once, 2 more per second
	l.now = func() time.Time { return now }
	for i := 0; i < 3; i++ {
		if !l.Allow(7) {
			t.Fatalf("burst %d", i)
		}
	}
	if l.Allow(7) {
		t.Fatal("over the burst")
	}
	if !l.Allow(8) {
		t.Fatal("another user is unaffected")
	}
	now = now.Add(500 * time.Millisecond) // one token back
	if !l.Allow(7) || l.Allow(7) {
		t.Fatal("refill")
	}
	now = now.Add(time.Hour) // never more than the burst
	for i := 0; i < 3; i++ {
		if !l.Allow(7) {
			t.Fatal("refilled")
		}
	}
	if l.Allow(7) {
		t.Fatal("burst cap")
	}
	var none *UserLimiter
	if !none.Allow(1) || NewUserLimiter(0, 0) != nil {
		t.Fatal("disabled")
	}
}
