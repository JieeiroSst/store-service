package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func TestSoldOutCache(t *testing.T) {
	clk := &clock{t: time.Unix(1_000_000, 0)}
	c := NewSoldOutCache(2 * time.Second)
	c.now = clk.now

	if c.SoldOut(1, 1) {
		t.Fatal("an unknown type is not sold out")
	}
	c.MarkSoldOut(1, 3) // 3 tickets were refused
	if !c.SoldOut(1, 3) || !c.SoldOut(1, 5) {
		t.Fatal("asking for as many or more than was refused must be sold out")
	}
	if c.SoldOut(1, 2) {
		t.Fatal("asking for fewer than was refused must still reach the database")
	}
	if c.SoldOut(2, 3) {
		t.Fatal("another type is unaffected")
	}
	c.MarkSoldOut(1, 5) // a bigger request being refused must not raise the bar
	if !c.SoldOut(1, 3) {
		t.Fatal("the smaller known-sold-out quantity was forgotten")
	}
	clk.advance(3 * time.Second)
	if c.SoldOut(1, 3) {
		t.Fatal("sold out must expire")
	}

	c.MarkSoldOut(1, 1)
	c.Release(1) // tickets came back
	if c.SoldOut(1, 1) {
		t.Fatal("Release must invalidate at once")
	}

	c.MarkAvailable(1, 10)
	if !c.RecentlyAvailable(1, 10) || c.RecentlyAvailable(1, 11) {
		t.Fatal("recently-available must cover only what was seen free")
	}
	clk.advance(availableFor + time.Millisecond)
	if c.RecentlyAvailable(1, 1) {
		t.Fatal("recently-available must be short-lived")
	}

	var off *SoldOutCache // nil and ttl 0 mean "no cache"
	if off.SoldOut(1, 1) || off.RecentlyAvailable(1, 1) {
		t.Fatal("a nil cache must say nothing")
	}
	off.MarkSoldOut(1, 1)
	off.Release(1)
	if NewSoldOutCache(0).SoldOut(1, 1) {
		t.Fatal("ttl 0 disables the cache")
	}
}

func TestSoldOutCacheIsSafeUnderRace(t *testing.T) {
	c := NewSoldOutCache(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 500; n++ {
				id := int64(n % 7)
				switch (n + i) % 5 {
				case 0:
					c.MarkSoldOut(id, 1+n%3)
				case 1:
					c.MarkAvailable(id, n%9)
				case 2:
					c.Release(id)
				default:
					c.SoldOut(id, 2)
					c.RecentlyAvailable(id, 1)
				}
			}
		}()
	}
	wg.Wait()
}

func TestBulkheadAdmitsOnlyItsSizeAndThenRefuses(t *testing.T) {
	b := NewBulkhead(2, 20*time.Millisecond)
	ctx := context.Background()
	l1, err := b.Enter(ctx)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := b.Enter(ctx)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if _, err := b.Enter(ctx); !errors.Is(err, domain.ErrBusy) {
		t.Fatalf("a full bulkhead must answer ErrBusy, got %v", err)
	}
	if time.Since(start) < 15*time.Millisecond {
		t.Fatal("a full bulkhead should make the caller wait a little first")
	}
	l1()
	l3, err := b.Enter(ctx)
	if err != nil {
		t.Fatalf("a freed slot must be usable: %v", err)
	}
	l2()
	l3()

	cctx, cancel := context.WithCancel(ctx)
	cancel()
	l1, _ = b.Enter(ctx)
	l2, _ = b.Enter(ctx)
	if _, err := b.Enter(cctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("a cancelled caller must give up: %v", err)
	}
	l1()
	l2()

	var none *Bulkhead
	if leave, err := none.Enter(ctx); err != nil {
		t.Fatal(err)
	} else {
		leave()
	}
	if NewBulkhead(0, time.Second) != nil {
		t.Fatal("size 0 means no limit")
	}
}

func TestBulkheadNeverExceedsItsSize(t *testing.T) {
	b := NewBulkhead(4, time.Second)
	var inside, peak atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			leave, err := b.Enter(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
			n := inside.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			inside.Add(-1)
			leave()
		}()
	}
	wg.Wait()
	if peak.Load() > 4 {
		t.Fatalf("%d requests were inside a bulkhead of 4", peak.Load())
	}
}

func TestUserLimiter(t *testing.T) {
	clk := &clock{t: time.Unix(1_000_000, 0)}
	l := NewUserLimiter(2, 3) // 2 per second, burst 3
	l.now = clk.now

	for i := 0; i < 3; i++ {
		if !l.Allow(1) {
			t.Fatalf("attempt %d of the burst was refused", i+1)
		}
	}
	if l.Allow(1) {
		t.Fatal("the burst is used up")
	}
	if !l.Allow(2) {
		t.Fatal("another user has their own bucket")
	}
	clk.advance(500 * time.Millisecond) // one token
	if !l.Allow(1) || l.Allow(1) {
		t.Fatal("half a second at 2/s is exactly one token")
	}
	clk.advance(time.Hour)
	for i := 0; i < 3; i++ {
		if !l.Allow(1) {
			t.Fatal("the bucket must refill up to the burst")
		}
	}
	if l.Allow(1) {
		t.Fatal("...and no further")
	}

	var off *UserLimiter
	if !off.Allow(1) || NewUserLimiter(0, 5) != nil || NewUserLimiter(5, 0) != nil {
		t.Fatal("a zero rate or burst means no limit")
	}
}

func TestUserLimiterIsSafeUnderRace(t *testing.T) {
	l := NewUserLimiter(1000, 1000)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				l.Allow(int64(n % 50))
			}
		}()
	}
	wg.Wait()
}

func TestFlightsCollapseConcurrentLookups(t *testing.T) {
	f := newFlights()
	var calls atomic.Int32
	release := make(chan struct{})
	var wg sync.WaitGroup
	results := make([]int, 50)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := f.Do(7, func() (int, error) {
				calls.Add(1)
				<-release
				return 42, nil
			})
			if err != nil {
				t.Error(err)
			}
			results[i] = v
		}()
	}
	time.Sleep(50 * time.Millisecond) // let everyone queue behind the one call
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("%d lookups for 50 concurrent askers, want 1", calls.Load())
	}
	for _, v := range results {
		if v != 42 {
			t.Fatalf("results %v", results)
		}
	}
	// once done, the next asker starts a new lookup
	if v, _ := f.Do(7, func() (int, error) { return 43, nil }); v != 43 {
		t.Fatal("a finished flight must not be reused")
	}
	// errors are shared too, and do not stick
	boom := errors.New("boom")
	if _, err := f.Do(8, func() (int, error) { return 0, boom }); !errors.Is(err, boom) {
		t.Fatal("the error must come back")
	}
	if v, err := f.Do(8, func() (int, error) { return 1, nil }); v != 1 || err != nil {
		t.Fatal("a failed flight must not stick")
	}
}

func TestTTLCache(t *testing.T) {
	clk := &clock{t: time.Unix(1_000_000, 0)}
	c := newTTLCache[int, string](time.Second, 100)
	c.now = clk.now

	var loads atomic.Int32
	load := func() (string, error) { loads.Add(1); return "v", nil }
	for i := 0; i < 5; i++ {
		if v, err := c.Get(1, load); v != "v" || err != nil {
			t.Fatal(v, err)
		}
	}
	if loads.Load() != 1 {
		t.Fatalf("%d loads for 5 reads inside the TTL", loads.Load())
	}
	clk.advance(2 * time.Second)
	c.Get(1, load)
	if loads.Load() != 2 {
		t.Fatal("an expired entry must be loaded again")
	}
	c.Forget(1)
	c.Get(1, load)
	if loads.Load() != 3 {
		t.Fatal("a forgotten entry must be loaded again")
	}

	boom := errors.New("boom")
	if _, err := c.Get(2, func() (string, error) { return "", boom }); !errors.Is(err, boom) {
		t.Fatal("the load error must come back")
	}
	if v, _ := c.Get(2, func() (string, error) { return "ok", nil }); v != "ok" {
		t.Fatal("errors must not be cached")
	}

	off := newTTLCache[int, string](0, 10)
	off.Get(1, load)
	off.Get(1, load)
	if loads.Load() != 5 {
		t.Fatal("ttl 0 must load every time")
	}
	var nilCache *ttlCache[int, string]
	if v, _ := nilCache.Get(1, load); v != "v" {
		t.Fatal("a nil cache must just load")
	}
}

func TestTTLCacheCollapsesConcurrentMisses(t *testing.T) {
	c := newTTLCache[int, int](time.Minute, 100)
	var loads atomic.Int32
	release := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := c.Get(1, func() (int, error) { loads.Add(1); <-release; return 9, nil })
			if v != 9 || err != nil {
				t.Error(v, err)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	if loads.Load() != 1 {
		t.Fatalf("%d loads for 100 concurrent misses", loads.Load())
	}
}

func TestTicketCodesAreLongAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 10_000; i++ {
		c, err := newTicketCode()
		if err != nil {
			t.Fatal(err)
		}
		if len(c) != 32 || seen[c] {
			t.Fatalf("bad or repeated code %q", c)
		}
		seen[c] = true
	}
}
