package trie

import "sync"

type node struct {
	children map[rune]*node
	isEnd    bool
	value    interface{}
}

type Trie struct {
	mu   sync.RWMutex
	root *node
}

func New() *Trie { return &Trie{root: &node{children: make(map[rune]*node)}} }

func (t *Trie) Insert(word string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.root
	for _, ch := range word {
		if _, ok := cur.children[ch]; !ok {
			cur.children[ch] = &node{children: make(map[rune]*node)}
		}
		cur = cur.children[ch]
	}
	cur.isEnd = true
	cur.value = value
}

func (t *Trie) Search(word string) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	cur := t.root
	for _, ch := range word {
		if next, ok := cur.children[ch]; ok { cur = next } else { return nil, false }
	}
	if cur.isEnd { return cur.value, true }
	return nil, false
}

func (t *Trie) StartsWith(prefix string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	cur := t.root
	for _, ch := range prefix {
		if next, ok := cur.children[ch]; ok { cur = next } else { return false }
	}
	return true
}
