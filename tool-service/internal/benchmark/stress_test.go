package benchmark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStressFindsBreakingPoint(t *testing.T) {
	sem := make(chan struct{}, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sem <- struct{}{}
		defer func() { <-sem }()
		time.Sleep(30 * time.Millisecond)
	}))
	defer srv.Close()

	rep := Stress(context.Background(), srv.Client(), Config{URL: srv.URL}, []int{1, 2, 16, 32}, 1, 100, 0.05, nil)
	if rep.BreakingConcurrency != 16 || rep.MaxSustainable != 2 {
		t.Fatalf("breaking=%d sustainable=%d stages=%+v", rep.BreakingConcurrency, rep.MaxSustainable, rep.Stages)
	}
	if len(rep.Stages) != 3 || rep.PeakRPS <= 0 {
		t.Fatalf("must stop at the first violation: %+v", rep)
	}
	if v := rep.Stages[2].Violations; len(v) == 0 {
		t.Fatal("violations not reported")
	}
}

func TestSoakDetectsDegradation(t *testing.T) {
	var slow atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slow.Load() {
			time.Sleep(80 * time.Millisecond)
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer srv.Close()
	go func() { time.Sleep(6 * time.Second); slow.Store(true) }()

	rep := SoakWindows(context.Background(), srv.Client(), Config{URL: srv.URL, Concurrency: 2}, 12, 1)
	if !rep.Degraded || rep.Drift.P95ChangePct < 100 {
		t.Fatalf("degradation not detected: %+v", rep)
	}
}

func TestSoakStableIsNotFlagged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(5 * time.Millisecond) }))
	defer srv.Close()
	if rep := SoakWindows(context.Background(), srv.Client(), Config{URL: srv.URL, Concurrency: 2}, 8, 1); rep.Degraded {
		t.Fatalf("stable service flagged: %+v", rep)
	}
}
