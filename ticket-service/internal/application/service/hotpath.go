package service

import (
	"context"
	"sync"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

//	per-user rate limit   -> one account cannot hammer
//	sold-out cache        -> once a type has run out, the rest are refused from memory, no I/O at all
//	single flight         -> the "how many are left?" question is asked once for all concurrent askers
//	bulkhead              -> only a few requests per pod compete in the database, the rest queue briefly or get 429

// ---- sold-out cache

type soldOutEntry struct {
	qty     int
	gen     uint64
	expires time.Time
}

type availEntry struct {
	qty     int
	gen     uint64
	expires time.Time
}

type SoldOutCache struct {
	ttl     time.Duration
	max     int
	now     func() time.Time
	mu      sync.Mutex
	entries map[int64]soldOutEntry
	avail   map[int64]availEntry
	gen     map[int64]uint64
}

const availableFor = 50 * time.Millisecond

func NewSoldOutCache(ttl time.Duration) *SoldOutCache {
	return &SoldOutCache{ttl: ttl, max: 100_000, now: time.Now,
		entries: map[int64]soldOutEntry{}, avail: map[int64]availEntry{}, gen: map[int64]uint64{}}
}

func (c *SoldOutCache) SoldOut(typeID int64, qty int) bool {
	if c == nil || c.ttl <= 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[typeID]
	if !ok {
		return false
	}
	if c.now().After(e.expires) || e.gen != c.gen[typeID] {
		delete(c.entries, typeID)
		return false
	}
	return qty >= e.qty
}

func (c *SoldOutCache) MarkSoldOut(typeID int64, qty int) {
	if c == nil || c.ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		clear(c.entries)
	}
	if e, ok := c.entries[typeID]; ok && e.gen == c.gen[typeID] && !c.now().After(e.expires) && e.qty < qty {
		qty = e.qty
	}
	c.entries[typeID] = soldOutEntry{qty: qty, gen: c.gen[typeID], expires: c.now().Add(c.ttl)}
	delete(c.avail, typeID)
}

func (c *SoldOutCache) RecentlyAvailable(typeID int64, qty int) bool {
	if c == nil || c.ttl <= 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.avail[typeID]
	if !ok || c.now().After(e.expires) || e.gen != c.gen[typeID] {
		return false
	}
	return qty <= e.qty
}

func (c *SoldOutCache) MarkAvailable(typeID int64, qty int) {
	if c == nil || c.ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.avail) >= c.max {
		clear(c.avail)
	}
	c.avail[typeID] = availEntry{qty: qty, gen: c.gen[typeID], expires: c.now().Add(availableFor)}
}

func (c *SoldOutCache) Release(typeID int64) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.gen[typeID]++
	c.mu.Unlock()
}

// ---- bulkhead

type Bulkhead struct {
	slots chan struct{}
	wait  time.Duration
}

func NewBulkhead(size int, wait time.Duration) *Bulkhead {
	if size <= 0 {
		return nil
	}
	return &Bulkhead{slots: make(chan struct{}, size), wait: wait}
}

func (b *Bulkhead) Enter(ctx context.Context) (func(), error) {
	if b == nil {
		return func() {}, nil
	}
	select {
	case b.slots <- struct{}{}:
		return func() { <-b.slots }, nil
	default:
	}
	t := time.NewTimer(b.wait)
	defer t.Stop()
	select {
	case b.slots <- struct{}{}:
		return func() { <-b.slots }, nil
	case <-t.C:
		return nil, domain.ErrBusy
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ---- per-user rate limit

type UserLimiter struct {
	perSecond float64
	burst     float64
	now       func() time.Time

	mu      sync.Mutex
	buckets map[int64]*bucket
	max     int
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewUserLimiter(perSecond float64, burst int) *UserLimiter {
	if perSecond <= 0 || burst <= 0 {
		return nil
	}
	return &UserLimiter{perSecond: perSecond, burst: float64(burst), now: time.Now, buckets: map[int64]*bucket{}, max: 500_000}
}

func (l *UserLimiter) Allow(user int64) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[user]
	if !ok {
		if len(l.buckets) >= l.max {
			for id, x := range l.buckets {
				if x.tokens+now.Sub(x.last).Seconds()*l.perSecond >= l.burst {
					delete(l.buckets, id) // full buckets are indistinguishable from new ones
				}
			}
			if len(l.buckets) >= l.max {
				clear(l.buckets)
			}
		}
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[user] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.perSecond)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// ---- single flight

type flights struct {
	mu    sync.Mutex
	calls map[int64]*flightCall
}

type flightCall struct {
	done chan struct{}
	val  int
	err  error
}

func newFlights() *flights { return &flights{calls: map[int64]*flightCall{}} }

func (f *flights) Do(key int64, fn func() (int, error)) (int, error) {
	f.mu.Lock()
	if c, ok := f.calls[key]; ok {
		f.mu.Unlock()
		<-c.done
		return c.val, c.err
	}
	c := &flightCall{done: make(chan struct{})}
	f.calls[key] = c
	f.mu.Unlock()

	c.val, c.err = fn()
	f.mu.Lock()
	delete(f.calls, key)
	f.mu.Unlock()
	close(c.done)
	return c.val, c.err
}
