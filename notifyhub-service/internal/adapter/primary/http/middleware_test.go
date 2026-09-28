package http

import (
	"testing"
	"time"
)

func TestRateLimiterWindowResetsUnderSteadyTraffic(t *testing.T) {
	rl := NewRateLimiter(2)
	defer rl.Close()
	start := time.Unix(0, 0)

	if !rl.Allow("ip", start) || !rl.Allow("ip", start.Add(100*time.Millisecond)) {
		t.Fatal("first two requests must pass")
	}
	if rl.Allow("ip", start.Add(900*time.Millisecond)) {
		t.Fatal("third request in the same window must be limited")
	}
	// A client that keeps sending must be allowed again in the next window.
	if !rl.Allow("ip", start.Add(1100*time.Millisecond)) {
		t.Fatal("request in the next window must pass")
	}
	if !rl.Allow("other", start) {
		t.Fatal("limits are per client")
	}
}
