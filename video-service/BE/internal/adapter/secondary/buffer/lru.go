package buffer

import (
	"container/list"
	"hash/fnv"
	"sync"
)

const shardCount = 16

type chunkKey struct {
	object string
	index  int64
}

type lru struct {
	shards [shardCount]*shard
}

type shard struct {
	mu       sync.Mutex
	maxBytes int64
	bytes    int64
	items    map[chunkKey]*list.Element
	order    *list.List
}

type entry struct {
	key  chunkKey
	data []byte
}

func newLRU(maxBytes int64) *lru {
	c := &lru{}
	for i := range c.shards {
		c.shards[i] = &shard{
			maxBytes: max(maxBytes/shardCount, 1),
			items:    map[chunkKey]*list.Element{},
			order:    list.New(),
		}
	}
	return c
}

func (c *lru) shard(k chunkKey) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(k.object))
	return c.shards[(h.Sum32()+uint32(k.index))%shardCount]
}

func (c *lru) get(k chunkKey) ([]byte, bool) {
	s := c.shard(k)
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.items[k]
	if !ok {
		return nil, false
	}
	s.order.MoveToFront(el)
	return el.Value.(*entry).data, true
}

func (c *lru) contains(k chunkKey) bool {
	s := c.shard(k)
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[k]
	return ok
}

func (c *lru) add(k chunkKey, data []byte) {
	s := c.shard(k)
	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.items[k]; ok {
		s.order.MoveToFront(el)
		return
	}
	s.items[k] = s.order.PushFront(&entry{key: k, data: data})
	s.bytes += int64(len(data))
	for s.bytes > s.maxBytes && s.order.Len() > 1 {
		last := s.order.Back()
		e := last.Value.(*entry)
		s.order.Remove(last)
		delete(s.items, e.key)
		s.bytes -= int64(len(e.data))
	}
}
