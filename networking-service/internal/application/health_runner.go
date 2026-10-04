package application

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type HealthRunner struct {
	store    port.CatalogStore
	prober   port.Prober
	metrics  port.Metrics
	leader   port.Leadership
	sessions *SessionService
	tick     time.Duration
	timeout  time.Duration
	now      func() time.Time

	sem     chan struct{}
	mu      sync.Mutex
	next    map[string]time.Time
	running map[string]bool
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

func NewHealthRunner(store port.CatalogStore, prober port.Prober, metrics port.Metrics, leader port.Leadership, sessions *SessionService, cfg *config.Config) *HealthRunner {
	return &HealthRunner{
		store:    store,
		prober:   prober,
		metrics:  metrics,
		leader:   leader,
		sessions: sessions,
		tick:     cfg.Health.TickInterval,
		timeout:  cfg.Health.DefaultTimeout,
		now:      time.Now,
		sem:      make(chan struct{}, cfg.Health.MaxConcurrent),
		next:     map[string]time.Time{},
		running:  map[string]bool{},
	}
}

func (r *HealthRunner) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		t := time.NewTicker(r.tick)
		defer t.Stop()
		for {
			if err := r.Tick(ctx); err != nil {
				log.Printf("health: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

func (r *HealthRunner) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.wg.Wait()
}

func (r *HealthRunner) Tick(ctx context.Context) error {
	if !r.leader.IsLeader() {
		r.mu.Lock()
		clear(r.next)
		r.mu.Unlock()
		return nil
	}
	now := r.now()
	_, checks, err := r.store.Checks()
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(checks))
	for _, c := range checks {
		key := c.Node + "/" + c.ID
		seen[key] = true

		switch c.Type {
		case domain.CheckHTTP, domain.CheckTCP:
			r.schedule(ctx, key, c, now)
		case domain.CheckTTL:
			r.expireTTL(c, now)
		}

		if c.ServiceID != "" && c.DeregisterCriticalServiceAfter > 0 && c.Status == domain.HealthCritical &&
			!c.CriticalSince.IsZero() && now.Sub(c.CriticalSince) > c.DeregisterCriticalServiceAfter {
			log.Printf("health: deregistering %s/%s, critical since %s", c.Node, c.ServiceID, c.CriticalSince.Format(time.RFC3339))
			_ = r.store.DeregisterService(c.Node, c.ServiceID)
		}
	}

	r.mu.Lock()
	for key := range r.next {
		if !seen[key] && !r.running[key] {
			delete(r.next, key)
		}
	}
	r.mu.Unlock()

	return r.sessions.ReapExpired()
}

func (r *HealthRunner) schedule(ctx context.Context, key string, c domain.Check, now time.Time) {
	r.mu.Lock()
	if r.running[key] || now.Before(r.next[key]) {
		r.mu.Unlock()
		return
	}
	select {
	case r.sem <- struct{}{}:
	default:
		r.mu.Unlock()
		return
	}
	r.running[key] = true
	r.next[key] = now.Add(c.Interval)
	r.mu.Unlock()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer func() {
			<-r.sem
			r.mu.Lock()
			delete(r.running, key)
			r.mu.Unlock()
		}()
		timeout := c.Timeout
		if timeout <= 0 {
			timeout = r.timeout
		}
		pctx, cancel := context.WithTimeout(ctx, timeout)
		status, output := r.prober.Probe(pctx, c)
		cancel()
		if ctx.Err() != nil {
			return
		}
		r.metrics.CheckRun(c.Type, status)
		_, _ = r.store.UpdateCheck(c.Node, c.ID, status, output, time.Time{})
	}()
}

func (r *HealthRunner) expireTTL(c domain.Check, now time.Time) {
	if c.TTLExpires.IsZero() {
		_, _ = r.store.UpdateCheck(c.Node, c.ID, c.Status, c.Output, now.Add(c.TTL))
		return
	}
	if c.Status == domain.HealthCritical || now.Before(c.TTLExpires) {
		return
	}
	output := "TTL expired"
	if c.Output != "" {
		output += " (last output before timeout follows): " + c.Output
	}
	r.metrics.CheckRun(c.Type, domain.HealthCritical)
	_, _ = r.store.UpdateCheck(c.Node, c.ID, domain.HealthCritical, output, c.TTLExpires)
}
