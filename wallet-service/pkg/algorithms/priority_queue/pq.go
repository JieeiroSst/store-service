package priority_queue

import "container/heap"

type Item[T any] struct {
	Value    T
	Priority int
	index    int
}

type inner[T any] []*Item[T]

func (h inner[T]) Len() int           { return len(h) }
func (h inner[T]) Less(i, j int) bool { return h[i].Priority < h[j].Priority }
func (h inner[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *inner[T]) Push(x interface{}) {
	i := x.(*Item[T]); i.index = len(*h); *h = append(*h, i)
}
func (h *inner[T]) Pop() interface{} {
	old := *h; n := len(old); i := old[n-1]; old[n-1] = nil; *h = old[:n-1]; return i
}

type PriorityQueue[T any] struct{ h *inner[T] }

func New[T any]() *PriorityQueue[T] {
	h := &inner[T]{}; heap.Init(h); return &PriorityQueue[T]{h: h}
}

func (pq *PriorityQueue[T]) Push(value T, priority int) {
	heap.Push(pq.h, &Item[T]{Value: value, Priority: priority})
}

func (pq *PriorityQueue[T]) Pop() (T, bool) {
	if pq.h.Len() == 0 { var z T; return z, false }
	return heap.Pop(pq.h).(*Item[T]).Value, true
}

func (pq *PriorityQueue[T]) Peek() (T, bool) {
	if pq.h.Len() == 0 { var z T; return z, false }
	return (*pq.h)[0].Value, true
}

func (pq *PriorityQueue[T]) Len() int { return pq.h.Len() }
