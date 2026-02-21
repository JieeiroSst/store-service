package lru

import (
	"container/list"
	"sync"
)

type Cache[K comparable, V any] struct {
	cap   int
	mu    sync.RWMutex
	lst   *list.List
	items map[K]*list.Element
}

type entry[K comparable, V any] struct{ key K; value V }

func New[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 { capacity = 128 }
	return &Cache[K, V]{cap: capacity, lst: list.New(), items: make(map[K]*list.Element, capacity)}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.lst.MoveToFront(elem)
		return elem.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (c *Cache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.lst.MoveToFront(elem)
		elem.Value.(*entry[K, V]).value = value
		return
	}
	if c.lst.Len() >= c.cap {
		if tail := c.lst.Back(); tail != nil {
			c.lst.Remove(tail)
			delete(c.items, tail.Value.(*entry[K, V]).key)
		}
	}
	e := &entry[K, V]{key: key, value: value}
	elem := c.lst.PushFront(e)
	c.items[key] = elem
}

func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.lst.Remove(elem)
		delete(c.items, key)
		return true
	}
	return false
}

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lst.Len()
}
