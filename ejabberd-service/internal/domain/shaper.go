package domain

import (
	"sync"
	"time"
)

type ShaperSpec struct {
	Rate  int
	Burst int
}

type Shaper struct {
	mu       sync.Mutex
	rate     float64
	burst    float64
	tokens   float64
	lastTime time.Time
	now      func() time.Time
}

func NewShaper(spec ShaperSpec) *Shaper {
	if spec.Rate <= 0 {
		return nil
	}
	burst := spec.Burst
	if burst <= 0 {
		burst = spec.Rate
	}
	return &Shaper{rate: float64(spec.Rate), burst: float64(burst), tokens: float64(burst), lastTime: time.Now(), now: time.Now}
}

func (s *Shaper) Update(size int) time.Duration {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.tokens += now.Sub(s.lastTime).Seconds() * s.rate
	if s.tokens > s.burst {
		s.tokens = s.burst
	}
	s.lastTime = now
	s.tokens -= float64(size)
	if s.tokens >= 0 {
		return 0
	}
	return time.Duration(-s.tokens / s.rate * float64(time.Second))
}
