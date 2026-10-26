package breaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JIeeiroSst/utils/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrOpen = errors.New("circuit breaker is open")

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return "closed"
}

type Settings struct {
	Disabled         bool
	FailureThreshold int
	OpenTimeout      time.Duration
	HalfOpenRequests int
}

type Breaker struct {
	name     string
	settings Settings
	now      func() time.Time

	mu       sync.Mutex
	state    State
	failures int
	inflight int
	openedAt time.Time
}

func New(name string, s Settings) *Breaker {
	if s.FailureThreshold <= 0 {
		s.FailureThreshold = 5
	}
	if s.OpenTimeout <= 0 {
		s.OpenTimeout = 30 * time.Second
	}
	if s.HalfOpenRequests <= 0 {
		s.HalfOpenRequests = 1
	}
	return &Breaker{name: name, settings: s, now: time.Now}
}

func (b *Breaker) Name() string { return b.name }

func (b *Breaker) State() State {
	if b == nil {
		return Closed
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refresh()
	return b.state
}

func (b *Breaker) Allow() (func(failed bool), error) {
	if b == nil || b.settings.Disabled {
		return func(bool) {}, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refresh()
	switch b.state {
	case Open:
		return nil, ErrOpen
	case HalfOpen:
		if b.inflight >= b.settings.HalfOpenRequests {
			return nil, ErrOpen
		}
	}
	b.inflight++
	var once sync.Once
	return func(failed bool) {
		once.Do(func() { b.done(failed) })
	}, nil
}

func (b *Breaker) done(failed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inflight--
	if !failed {
		b.failures = 0
		if b.state == HalfOpen {
			b.transition(Closed)
		}
		return
	}
	b.failures++
	if b.state == HalfOpen || b.failures >= b.settings.FailureThreshold {
		b.openedAt = b.now()
		b.transition(Open)
	}
}

func (b *Breaker) refresh() {
	if b.state == Open && b.now().Sub(b.openedAt) >= b.settings.OpenTimeout {
		b.transition(HalfOpen)
	}
}

func (b *Breaker) transition(to State) {
	if b.state == to {
		return
	}
	logger.Info(context.Background(), "circuit breaker %s: %s -> %s", b.name, b.state, to)
	b.state = to
	if to != Open {
		b.failures = 0
	}
}

func GRPCFailure(err error) bool {
	if err == nil {
		return false
	}
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.DataLoss:
		return true
	}
	return false
}

func UnaryClientInterceptor(b *Breaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		done, err := b.Allow()
		if err != nil {
			return status.Error(codes.Unavailable, err.Error())
		}
		err = invoker(ctx, method, req, reply, cc, opts...)
		done(GRPCFailure(err) && !errors.Is(ctx.Err(), context.Canceled))
		return err
	}
}

func StreamClientInterceptor(b *Breaker) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		done, err := b.Allow()
		if err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		cs, err := streamer(ctx, desc, cc, method, opts...)
		done(GRPCFailure(err) && !errors.Is(ctx.Err(), context.Canceled))
		return cs, err
	}
}

func DialOptions(b *Breaker) []grpc.DialOption {
	if b == nil {
		return nil
	}
	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(UnaryClientInterceptor(b)),
		grpc.WithChainStreamInterceptor(StreamClientInterceptor(b)),
	}
}
