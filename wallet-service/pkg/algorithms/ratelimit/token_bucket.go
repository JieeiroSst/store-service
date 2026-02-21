package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu           sync.Mutex
	capacity     float64
	tokens       float64
	refillRate   float64 
	lastRefill   time.Time
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{capacity: capacity, tokens: capacity, refillRate: refillRate, lastRefill: time.Now()}
}

func (tb *TokenBucket) Allow() bool { return tb.AllowN(1) }

func (tb *TokenBucket) AllowN(n float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min64(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now
	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

func min64(a, b float64) float64 {
	if a < b { return a }
	return b
}

type SlidingWindowLimiter struct {
	mu         sync.Mutex
	limit      int
	window     time.Duration
	requests   []time.Time
}

func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{limit: limit, window: window, requests: make([]time.Time, 0, limit)}
}

func (sw *SlidingWindowLimiter) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-sw.window)
	valid := sw.requests[:0]
	for _, t := range sw.requests {
		if t.After(cutoff) { valid = append(valid, t) }
	}
	sw.requests = valid
	if len(sw.requests) >= sw.limit { return false }
	sw.requests = append(sw.requests, now)
	return true
}
