package batch

import (
	"context"
	"sync"
	"time"
)

type ProcessFunc[T any] func(ctx context.Context, items []T) error

type Option[T any] func(*Processor[T])

func WithSize[T any](size int) Option[T]         { return func(p *Processor[T]) { p.batchSize = size } }
func WithFlushInterval[T any](d time.Duration) Option[T] { return func(p *Processor[T]) { p.interval = d } }
func WithProcessFunc[T any](fn ProcessFunc[T]) Option[T] { return func(p *Processor[T]) { p.fn = fn } }

type Processor[T any] struct {
	mu        sync.Mutex
	items     []T
	batchSize int
	interval  time.Duration
	fn        ProcessFunc[T]
	done      chan struct{}
}

func NewProcessor[T any](opts ...Option[T]) *Processor[T] {
	p := &Processor[T]{batchSize: 50, interval: 10 * time.Second, done: make(chan struct{})}
	for _, opt := range opts { opt(p) }
	return p
}

func (p *Processor[T]) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C: p.Flush(ctx)
		case <-ctx.Done(): p.Flush(ctx); return
		case <-p.done: return
		}
	}
}

func (p *Processor[T]) Add(ctx context.Context, item T) error {
	p.mu.Lock()
	p.items = append(p.items, item)
	should := len(p.items) >= p.batchSize
	p.mu.Unlock()
	if should { return p.Flush(ctx) }
	return nil
}

func (p *Processor[T]) Flush(ctx context.Context) error {
	p.mu.Lock()
	if len(p.items) == 0 { p.mu.Unlock(); return nil }
	batch := make([]T, len(p.items))
	copy(batch, p.items)
	p.items = p.items[:0]
	p.mu.Unlock()
	if p.fn != nil { return p.fn(ctx, batch) }
	return nil
}

func (p *Processor[T]) Len() int { p.mu.Lock(); defer p.mu.Unlock(); return len(p.items) }
func (p *Processor[T]) Stop()    { close(p.done) }
