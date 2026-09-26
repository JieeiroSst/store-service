package service

import (
	"sync"
	"time"
)

type ttlCache[K comparable, V any] struct {
	ttl time.Duration
	max int
	now func() time.Time

	mu      sync.Mutex
	entries map[K]cacheEntry[V]
	loading map[K]*cacheCall[V]
}

type cacheEntry[V any] struct {
	val     V
	expires time.Time
}

type cacheCall[V any] struct {
	done chan struct{}
	val  V
	err  error
}

func newTTLCache[K comparable, V any](ttl time.Duration, max int) *ttlCache[K, V] {
	return &ttlCache[K, V]{ttl: ttl, max: max, now: time.Now, entries: map[K]cacheEntry[V]{}, loading: map[K]*cacheCall[V]{}}
}

func (c *ttlCache[K, V]) Get(key K, load func() (V, error)) (V, error) {
	if c == nil || c.ttl <= 0 {
		return load()
	}
	c.mu.Lock()
	if e, ok := c.entries[key]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		return e.val, nil
	}
	if call, ok := c.loading[key]; ok {
		c.mu.Unlock()
		<-call.done
		return call.val, call.err
	}
	call := &cacheCall[V]{done: make(chan struct{})}
	c.loading[key] = call
	c.mu.Unlock()

	call.val, call.err = load()

	c.mu.Lock()
	delete(c.loading, key)
	if call.err == nil {
		if len(c.entries) >= c.max {
			now := c.now()
			for k, e := range c.entries {
				if !now.Before(e.expires) {
					delete(c.entries, k)
				}
			}
			if len(c.entries) >= c.max {
				clear(c.entries)
			}
		}
		c.entries[key] = cacheEntry[V]{val: call.val, expires: c.now().Add(c.ttl)}
	}
	c.mu.Unlock()
	close(call.done)
	return call.val, call.err
}

func (c *ttlCache[K, V]) Forget(key K) {
	if c == nil {
		return
	}
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}
