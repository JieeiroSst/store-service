package application

import (
	"context"
	"crypto/rand"
	"fmt"
	mrand "math/rand/v2"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type QueryOptions struct {
	MinIndex uint64
	Wait     time.Duration
	NodeMeta map[string]string
}

type QueryMeta struct {
	LastIndex uint64
}

type Blocker struct {
	defaultWait time.Duration
	maxWait     time.Duration
	metrics     port.Metrics
}

func NewBlocker(cfg *config.Config, metrics port.Metrics) *Blocker {
	return &Blocker{defaultWait: cfg.Blocking.DefaultWait, maxWait: cfg.Blocking.MaxWait, metrics: metrics}
}

func (b *Blocker) Query(ctx context.Context, w port.Watcher, q QueryOptions, read func() (uint64, error)) (QueryMeta, error) {
	if q.MinIndex == 0 {
		idx, err := read()
		return meta(idx), err
	}
	wait := q.Wait
	if wait <= 0 {
		wait = b.defaultWait
	}
	wait = min(wait, b.maxWait)
	wait += mrand.N(wait/16 + 1)
	timer := time.NewTimer(wait)
	defer timer.Stop()

	for {
		ch := w.WatchCh()
		idx, err := read()
		if err != nil {
			return QueryMeta{}, err
		}
		if idx > q.MinIndex {
			return meta(idx), nil
		}
		last, err := w.LastIndex()
		if err != nil {
			return QueryMeta{}, err
		}
		if q.MinIndex > last {
			return meta(idx), nil
		}
		select {
		case <-ch:
			b.metrics.BlockingQueryWoken()
		case <-timer.C:
			return meta(idx), nil
		case <-ctx.Done():
			return meta(idx), nil
		}
	}
}

func meta(idx uint64) QueryMeta {
	return QueryMeta{LastIndex: max(idx, 1)}
}

func matchNodeMeta(meta, want map[string]string) bool {
	for k, v := range want {
		if meta[k] != v {
			return false
		}
	}
	return true
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type SoloLeader struct{}

func (SoloLeader) IsLeader() bool { return true }
