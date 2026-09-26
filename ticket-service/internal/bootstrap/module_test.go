package bootstrap

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestStartWithRetryKeepsTryingUntilItWorks(t *testing.T) {
	calls := 0
	start := func() error {
		calls++
		if calls < 3 {
			return errors.New("temporal is down")
		}
		return nil
	}
	startWithRetry(context.Background(), start, time.Millisecond, quiet())
	if calls != 3 {
		t.Fatalf("%d attempts", calls)
	}
}

func TestStartWithRetryStopsWhenTheServiceStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		startWithRetry(ctx, func() error { return errors.New("down") }, time.Hour, quiet())
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("a stopping service must not wait for the next attempt")
	}
}
