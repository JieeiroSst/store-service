package worker

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/notifyhub-service/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Pool struct {
	mu     sync.RWMutex
	closed bool
	queue  chan port.Task
	size   int
	wg     sync.WaitGroup
	log    *zap.Logger
}

func NewPool(cfg *config.Config, log *zap.Logger) *Pool {
	return &Pool{
		queue: make(chan port.Task, cfg.Worker.QueueSize),
		size:  cfg.Worker.PoolSize,
		log:   log,
	}
}

func (p *Pool) Enqueue(t port.Task) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return port.ErrQueueClosed
	}
	select {
	case p.queue <- t:
		metrics.WorkerQueueDepth.Set(float64(len(p.queue)))
		return nil
	default:
		p.log.Warn("worker queue full, dropping task",
			zap.String("job_id", t.JobID),
			zap.Int("queue_cap", cap(p.queue)),
		)
		return port.ErrQueueFull
	}
}

func (p *Pool) start(dispatch port.DispatchUsecase) {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go func(id int) {
			defer p.wg.Done()
			for t := range p.queue {
				metrics.WorkerQueueDepth.Set(float64(len(p.queue)))
				p.run(dispatch, t, id)
			}
		}(i)
	}
	p.log.Info("worker pool started", zap.Int("workers", p.size), zap.Int("queue_size", cap(p.queue)))
}

func (p *Pool) run(dispatch port.DispatchUsecase, t port.Task, workerID int) {
	defer func() {
		if r := recover(); r != nil {
			p.log.Error("worker panic recovered", zap.String("job_id", t.JobID), zap.Any("panic", r))
		}
	}()
	if err := dispatch.Execute(context.Background(), t); err != nil {
		p.log.Warn("job run failed", zap.String("job_id", t.JobID), zap.Int("worker", workerID), zap.Error(err))
	}
}

func (p *Pool) stop(ctx context.Context) error {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.queue)
	}
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		p.log.Info("worker pool drained")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func RunPool(lc fx.Lifecycle, pool *Pool, dispatch port.DispatchUsecase) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			pool.start(dispatch)
			return nil
		},
		OnStop: pool.stop,
	})
}
