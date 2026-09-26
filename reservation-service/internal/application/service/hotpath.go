package service

import (
	"context"
	"sync"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// ---- sold-out cache

type soldOutKey struct {
	roomType   int64
	start, end time.Time
}

type soldOutEntry struct {
	rooms   int // requests for this many rooms or more are sold out
	gen     uint64
	expires time.Time
}

type SoldOutCache struct {
	ttl     time.Duration
	max     int
	now     func() time.Time
	mu      sync.Mutex
	entries map[soldOutKey]soldOutEntry
	gen     map[int64]uint64 // per room type; bumped by Release
	avail   map[soldOutKey]availEntry
}

type availEntry struct {
	rooms   int // this many rooms were seen free
	gen     uint64
	expires time.Time
}

const availableFor = 50 * time.Millisecond

func NewSoldOutCache(ttl time.Duration) *SoldOutCache {
	return &SoldOutCache{ttl: ttl, max: 100_000, now: time.Now, entries: map[soldOutKey]soldOutEntry{}, gen: map[int64]uint64{}, avail: map[soldOutKey]availEntry{}}
}

func (c *SoldOutCache) RecentlyAvailable(roomType int64, start, end time.Time, rooms int) bool {
	if c == nil || c.ttl <= 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.avail[c.key(roomType, start, end)]
	if !ok || c.now().After(e.expires) || e.gen != c.gen[roomType] {
		return false
	}
	return rooms <= e.rooms
}

func (c *SoldOutCache) MarkAvailable(roomType int64, start, end time.Time, rooms int) {
	if c == nil || c.ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.avail) >= c.max {
		clear(c.avail)
	}
	c.avail[c.key(roomType, start, end)] = availEntry{rooms: rooms, gen: c.gen[roomType], expires: c.now().Add(availableFor)}
}

func (c *SoldOutCache) key(roomType int64, start, end time.Time) soldOutKey {
	return soldOutKey{roomType, start, end}
}

func (c *SoldOutCache) SoldOut(roomType int64, start, end time.Time, rooms int) bool {
	if c == nil || c.ttl <= 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[c.key(roomType, start, end)]
	if !ok {
		return false
	}
	if c.now().After(e.expires) || e.gen != c.gen[roomType] {
		delete(c.entries, c.key(roomType, start, end))
		return false
	}
	return rooms >= e.rooms
}

func (c *SoldOutCache) MarkSoldOut(roomType int64, start, end time.Time, rooms int) {
	if c == nil || c.ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		now := c.now()
		for k, e := range c.entries {
			if now.After(e.expires) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= c.max {
			clear(c.entries)
		}
	}
	k := c.key(roomType, start, end)
	if e, ok := c.entries[k]; ok && e.gen == c.gen[roomType] && !c.now().After(e.expires) && e.rooms < rooms {
		rooms = e.rooms
	}
	c.entries[k] = soldOutEntry{rooms: rooms, gen: c.gen[roomType], expires: c.now().Add(c.ttl)}
	delete(c.avail, k)
}

func (c *SoldOutCache) Release(roomType int64) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.gen[roomType]++
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
					delete(l.buckets, id)
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
	calls map[soldOutKey]*flightCall
}

type flightCall struct {
	done chan struct{}
	val  int
	err  error
}

func newFlights() *flights { return &flights{calls: map[soldOutKey]*flightCall{}} }

func (f *flights) Do(k soldOutKey, fn func() (int, error)) (int, error) {
	f.mu.Lock()
	if c, ok := f.calls[k]; ok {
		f.mu.Unlock()
		<-c.done
		return c.val, c.err
	}
	c := &flightCall{done: make(chan struct{})}
	f.calls[k] = c
	f.mu.Unlock()

	c.val, c.err = fn()
	f.mu.Lock()
	delete(f.calls, k)
	f.mu.Unlock()
	close(c.done)
	return c.val, c.err
}
