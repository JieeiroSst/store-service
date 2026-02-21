package consistent_hash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
)

type Ring struct {
	mu           sync.RWMutex
	virtNodes    int
	nodes        map[uint32]string
	sortedHashes []uint32
}

func New(virtualNodes int) *Ring {
	if virtualNodes <= 0 { virtualNodes = 150 }
	return &Ring{virtNodes: virtualNodes, nodes: make(map[uint32]string)}
}

func (r *Ring) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < r.virtNodes; i++ {
		h := r.hash(fmt.Sprintf("%s#%d", node, i))
		r.nodes[h] = node
		r.sortedHashes = append(r.sortedHashes, h)
	}
	sort.Slice(r.sortedHashes, func(i, j int) bool { return r.sortedHashes[i] < r.sortedHashes[j] })
}

func (r *Ring) RemoveNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < r.virtNodes; i++ {
		delete(r.nodes, r.hash(fmt.Sprintf("%s#%d", node, i)))
	}
	r.sortedHashes = r.sortedHashes[:0]
	for h := range r.nodes { r.sortedHashes = append(r.sortedHashes, h) }
	sort.Slice(r.sortedHashes, func(i, j int) bool { return r.sortedHashes[i] < r.sortedHashes[j] })
}

func (r *Ring) GetNode(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sortedHashes) == 0 { return "", false }
	h := r.hash(key)
	idx := sort.Search(len(r.sortedHashes), func(i int) bool { return r.sortedHashes[i] >= h })
	if idx == len(r.sortedHashes) { idx = 0 }
	return r.nodes[r.sortedHashes[idx]], true
}

func (r *Ring) hash(key string) uint32 {
	sum := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(sum[:4])
}
