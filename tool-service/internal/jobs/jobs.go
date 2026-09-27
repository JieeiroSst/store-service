package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"
)

const (
	Running  = "running"
	Done     = "done"
	Failed   = "failed"
	Canceled = "canceled"
	keep     = 100
)

type Job struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	Status     string     `json:"status"`
	Progress   string     `json:"progress,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Result     any        `json:"result,omitempty"`
	Error      string     `json:"error,omitempty"`

	cancel context.CancelFunc
}

type Manager struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

func New() *Manager { return &Manager{jobs: map[string]*Job{}} }

type Func func(ctx context.Context, progress func(string)) (any, error)

func (m *Manager) Start(kind string, timeout time.Duration, fn Func) *Job {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	j := &Job{ID: hex.EncodeToString(b), Kind: kind, Status: Running, StartedAt: time.Now().UTC(), cancel: cancel}

	m.mu.Lock()
	m.jobs[j.ID] = j
	m.prune()
	m.mu.Unlock()

	go func() {
		defer cancel()
		res, err := fn(ctx, func(p string) { m.update(j, func() { j.Progress = p }) })
		m.update(j, func() {
			now := time.Now().UTC()
			j.FinishedAt = &now
			switch {
			case errors.Is(ctx.Err(), context.Canceled) && j.Status == Canceled:
			case err != nil:
				j.Status, j.Error = Failed, err.Error()
			default:
				j.Status, j.Result = Done, res
			}
		})
	}()
	return j
}

func (m *Manager) update(j *Job, f func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f()
}

func (m *Manager) Get(id string) (Job, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return Job{}, false
	}
	return *j, true
}

func (m *Manager) Cancel(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok || j.Status != Running {
		return false
	}
	j.Status = Canceled
	j.cancel()
	return true
}

func (m *Manager) List() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		c := *j
		c.Result = nil
		out = append(out, c)
	}
	sort.Slice(out, func(i, k int) bool { return out[i].StartedAt.After(out[k].StartedAt) })
	return out
}

func (m *Manager) prune() {
	if len(m.jobs) <= keep {
		return
	}
	var finished []*Job
	for _, j := range m.jobs {
		if j.Status != Running {
			finished = append(finished, j)
		}
	}
	sort.Slice(finished, func(i, k int) bool { return finished[i].StartedAt.Before(finished[k].StartedAt) })
	for i := 0; len(m.jobs) > keep && i < len(finished); i++ {
		delete(m.jobs, finished[i].ID)
	}
}
