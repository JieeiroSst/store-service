package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type Memory struct {
	mu         sync.Mutex
	maxEntries int
	ll         *list.List
	items      map[string]*list.Element
	now        func() time.Time
}

type entry struct {
	key     string
	value   []byte
	expires time.Time
}

func NewMemory(maxEntries int) *Memory {
	return &Memory{
		maxEntries: maxEntries,
		ll:         list.New(),
		items:      make(map[string]*list.Element),
		now:        time.Now,
	}
}

func (m *Memory) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	el, ok := m.items[key]
	if !ok {
		return nil, false, nil
	}
	e := el.Value.(*entry)
	if !m.now().Before(e.expires) {
		m.remove(el)
		return nil, false, nil
	}
	m.ll.MoveToFront(el)
	return e.value, true, nil
}

func (m *Memory) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	expires := m.now().Add(ttl)
	if el, ok := m.items[key]; ok {
		e := el.Value.(*entry)
		e.value, e.expires = value, expires
		m.ll.MoveToFront(el)
		return nil
	}
	m.items[key] = m.ll.PushFront(&entry{key: key, value: value, expires: expires})
	for m.ll.Len() > m.maxEntries {
		m.remove(m.ll.Back())
	}
	return nil
}

func (m *Memory) remove(el *list.Element) {
	m.ll.Remove(el)
	delete(m.items, el.Value.(*entry).key)
}

type None struct{}

func (None) Get(context.Context, string) ([]byte, bool, error)        { return nil, false, nil }
func (None) Set(context.Context, string, []byte, time.Duration) error { return nil }
