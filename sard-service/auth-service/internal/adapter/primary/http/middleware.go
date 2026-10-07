package http

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
)

type tokenSet [][]byte

func newTokenSet(tokens []string) tokenSet {
	var out tokenSet
	for _, t := range tokens {
		out = append(out, []byte(t))
	}
	return out
}

func (ts tokenSet) require(next http.HandlerFunc) http.HandlerFunc {
	if len(ts) == 0 {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if !ts.valid(bearer(r)) {
			fail(w, domain.ErrUnauthorized)
			return
		}
		next(w, r)
	}
}

func (ts tokenSet) valid(t string) bool {
	if t == "" {
		return false
	}
	ok := 0
	for _, want := range ts {
		ok |= subtle.ConstantTimeCompare([]byte(t), want)
	}
	return ok == 1
}

func bearer(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return r.Header.Get("X-Api-Token")
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

type limiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*bucket
	now     func() time.Time
}

type bucket struct {
	start time.Time
	count int
}

func newLimiter(limit int, window time.Duration) *limiter {
	return &limiter{limit: limit, window: window, buckets: map[string]*bucket{}, now: time.Now}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.buckets) > 10_000 {
		for k, b := range l.buckets {
			if now.Sub(b.start) >= l.window {
				delete(l.buckets, k)
			}
		}
	}
	b, ok := l.buckets[key]
	if !ok || now.Sub(b.start) >= l.window {
		l.buckets[key] = &bucket{start: now, count: 1}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	return true
}

func (l *limiter) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientKey(r)) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many attempts, try again later"})
			return
		}
		next(w, r)
	}
}

func clientKey(r *http.Request) string {
	if t := sessionToken(r); t != "" {
		sum := sha256.Sum256([]byte(t))
		return "s:" + hex.EncodeToString(sum[:8])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return "ip:" + host
}
