package breaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOpensAfterThresholdAndRecovers(t *testing.T) {
	now := time.Unix(0, 0)
	b := New("svc", Settings{FailureThreshold: 3, OpenTimeout: 10 * time.Second})
	b.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		done, err := b.Allow()
		if err != nil {
			t.Fatalf("call %d rejected while closed", i)
		}
		done(true)
	}
	if b.State() != Open {
		t.Fatalf("want open, got %s", b.State())
	}
	if _, err := b.Allow(); !errors.Is(err, ErrOpen) {
		t.Fatal("open breaker must reject calls")
	}

	now = now.Add(10 * time.Second)
	probe, err := b.Allow()
	if err != nil || b.State() != HalfOpen {
		t.Fatalf("want half-open probe, got %s %v", b.State(), err)
	}
	if _, err := b.Allow(); !errors.Is(err, ErrOpen) {
		t.Fatal("half-open breaker must limit concurrent probes")
	}
	probe(false)
	if b.State() != Closed {
		t.Fatalf("successful probe must close, got %s", b.State())
	}
}

func TestFailedProbeReopens(t *testing.T) {
	now := time.Unix(0, 0)
	b := New("svc", Settings{FailureThreshold: 1, OpenTimeout: time.Second})
	b.now = func() time.Time { return now }
	done, _ := b.Allow()
	done(true)
	now = now.Add(time.Second)
	probe, err := b.Allow()
	if err != nil {
		t.Fatal(err)
	}
	probe(true)
	if b.State() != Open {
		t.Fatalf("failed probe must reopen, got %s", b.State())
	}
}

func TestSuccessResetsFailures(t *testing.T) {
	b := New("svc", Settings{FailureThreshold: 2})
	for _, failed := range []bool{true, false, true, false, true} {
		done, err := b.Allow()
		if err != nil {
			t.Fatal("interleaved successes must keep the breaker closed")
		}
		done(failed)
	}
}

func TestDisabledAndNil(t *testing.T) {
	var nb *Breaker
	if _, err := nb.Allow(); err != nil {
		t.Fatal(err)
	}
	b := New("svc", Settings{Disabled: true, FailureThreshold: 1})
	for i := 0; i < 3; i++ {
		done, err := b.Allow()
		if err != nil {
			t.Fatal("disabled breaker must never reject")
		}
		done(true)
	}
}

func TestUnaryInterceptor(t *testing.T) {
	b := New("svc", Settings{FailureThreshold: 2})
	intercept := UnaryClientInterceptor(b)
	calls := 0
	invoke := func(code codes.Code) grpc.UnaryInvoker {
		return func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			calls++
			return status.Error(code, "x")
		}
	}
	for i := 0; i < 5; i++ {
		_ = intercept(context.Background(), "/m", nil, nil, nil, invoke(codes.NotFound))
	}
	if b.State() != Closed {
		t.Fatal("business errors must not trip the breaker")
	}
	for i := 0; i < 2; i++ {
		_ = intercept(context.Background(), "/m", nil, nil, nil, invoke(codes.Unavailable))
	}
	calls = 0
	err := intercept(context.Background(), "/m", nil, nil, nil, invoke(codes.OK))
	if status.Code(err) != codes.Unavailable || calls != 0 {
		t.Fatalf("open breaker must short-circuit with Unavailable, got %v calls=%d", err, calls)
	}
}
