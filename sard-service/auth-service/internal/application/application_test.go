package application

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
)

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

type cards struct{}

func (cards) ResolvePAN(_ context.Context, pan, _ string) (domain.CardRef, error) {
	if pan != "7301001234561234" {
		return domain.CardRef{}, domain.Conflict("card is not enrolled for payment authentication")
	}
	return domain.CardRef{CardID: "card-1", UserID: 7, MaskedPAN: "730100******1234", Status: "NORMAL"}, nil
}

type sessions struct{}

func (sessions) Validate(_ context.Context, token string) (int64, error) {
	switch token {
	case "alice":
		return 7, nil
	case "bob":
		return 8, nil
	}
	return 0, domain.ErrUnauthorized
}

type users struct{}

func (users) Username(_ context.Context, id int64) (string, error) {
	if id == 7 {
		return "anguyen", nil
	}
	return "", domain.Conflict("cardholder %d does not exist", id)
}

type otpProvider struct {
	mu     sync.Mutex
	clock  *clock
	n      int
	codes  map[string]string
	issued int
	limit  int
	down   bool
}

func (p *otpProvider) Issue(_ context.Context, username string) (string, time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.limit > 0 && p.issued >= p.limit {
		return "", time.Time{}, domain.ErrOTPLimit
	}
	p.issued++
	p.n++
	code := strconv.Itoa(100000 + p.n)
	p.codes[username] = code
	return code, p.clock.Now().Add(time.Minute), nil
}

func (p *otpProvider) Verify(_ context.Context, username, code string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.down {
		return false, domain.ErrUnavailable
	}
	return p.codes[username] == code, nil
}

type outbox struct {
	mu   sync.Mutex
	sent []port.OTPMessage
}

func (o *outbox) Send(_ context.Context, m port.OTPMessage) error {
	o.mu.Lock()
	o.sent = append(o.sent, m)
	o.mu.Unlock()
	return nil
}

func (o *outbox) last() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.sent[len(o.sent)-1].OTP
}

type nop struct{}

func (nop) ObserveChallenge(string) {}

var req = domain.InitiateRequest{PAN: "7301001234561234", Expiry: "10/31", Amount: 1_000_000, Currency: "VND", Merchant: "Shopee"}

func newService(expose bool) (*AuthenticationService, *outbox, *clock) {
	svc, box, clk, _ := newServiceWithProvider(expose)
	return svc, box, clk
}

func newServiceWithProvider(expose bool) (*AuthenticationService, *outbox, *clock, *otpProvider) {
	box := &outbox{}
	clk := &clock{t: time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)}
	provider := &otpProvider{clock: clk, codes: map[string]string{}}
	svc := NewAuthenticationService(Settings{
		Policy:    domain.Policy{OTPTTL: 5 * time.Minute, MaxAttempts: 3, MaxResends: 2, ValidityTime: 10 * time.Minute},
		ExposeOTP: expose,
	}, memory.NewChallenges(), memory.NewLocker(), cards{}, sessions{}, users{}, provider, box, clk, nop{})
	return svc, box, clk, provider
}

func TestHappyPath(t *testing.T) {
	svc, box, _ := newService(false)
	ctx := context.Background()
	out, err := svc.Initiate(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if out.OTP != "" {
		t.Fatal("otp must not be exposed unless configured")
	}
	code := box.last()
	if len(code) != domain.OTPLength || out.Challenge.Username != "anguyen" {
		t.Fatalf("otp %q challenge %+v", code, out.Challenge)
	}
	id := out.Challenge.ID

	pending, err := svc.ListPending(ctx, "alice")
	if err != nil || len(pending) != 1 || pending[0].ID != id {
		t.Fatalf("pending: %v %d", err, len(pending))
	}
	if list, _ := svc.ListPending(ctx, "bob"); len(list) != 0 {
		t.Fatal("bob must not see alice's challenges")
	}
	if _, err := svc.Verify(ctx, id, "bob", code); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("bob verifies: %v", err)
	}
	if _, err := svc.Verify(ctx, id, "nobody", code); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("bad session: %v", err)
	}
	c, err := svc.Verify(ctx, id, "alice", code)
	if err != nil || c.Status != domain.StatusAuthenticated {
		t.Fatalf("verify: %v %+v", err, c)
	}
	if _, err := svc.Consume(ctx, id, ConsumeCommand{CardID: "card-2", Amount: 1, Currency: "VND"}); !domain.IsConflict(err) {
		t.Fatalf("other card: %v", err)
	}
	if c, err := svc.Consume(ctx, id, ConsumeCommand{CardID: "card-1", Amount: 1_000_000, Currency: "VND"}); err != nil || c.Status != domain.StatusUsed {
		t.Fatalf("consume: %v", err)
	}
	if _, err := svc.Consume(ctx, id, ConsumeCommand{CardID: "card-1", Amount: 1, Currency: "VND"}); !domain.IsConflict(err) {
		t.Fatalf("replay: %v", err)
	}
}

func TestWrongOTPLocksAndPersists(t *testing.T) {
	svc, box, clk := newService(true)
	ctx := context.Background()
	out, _ := svc.Initiate(ctx, req)
	if out.OTP != box.last() {
		t.Fatal("expose mode returns the otp")
	}
	wrong := "000000"
	if out.OTP == wrong {
		wrong = "111111"
	}
	for i := 2; i >= 0; i-- {
		c, err := svc.Verify(ctx, out.Challenge.ID, "alice", wrong)
		var mm *domain.OTPMismatchError
		if !errors.As(err, &mm) || mm.Remaining != i || c == nil {
			t.Fatalf("attempt: %v %+v", err, c)
		}
	}
	got, _ := svc.Get(ctx, out.Challenge.ID)
	if got.Status != domain.StatusFailed || got.Attempts != 3 {
		t.Fatalf("not persisted: %+v", got)
	}
	if _, err := svc.Verify(ctx, out.Challenge.ID, "alice", out.OTP); !domain.IsConflict(err) {
		t.Fatalf("correct otp after lockout: %v", err)
	}

	second, _ := svc.Initiate(ctx, req)
	clk.advance(2 * time.Minute)
	if _, err := svc.Verify(ctx, second.Challenge.ID, "alice", second.OTP); !errors.Is(err, domain.ErrOTPExpired) {
		t.Fatalf("otp from authorize-service expired: %v", err)
	}
	resent, err := svc.Resend(ctx, second.Challenge.ID, "alice")
	if err != nil || resent.OTP == second.OTP || resent.Challenge.Resends != 1 {
		t.Fatalf("resend: %v %+v", err, resent)
	}
	if _, err := svc.Verify(ctx, second.Challenge.ID, "alice", second.OTP); !domain.IsOTPMismatch(err) {
		t.Fatalf("old code after resend: %v", err)
	}
	clk.advance(4 * time.Minute)
	if got, _ := svc.Get(ctx, second.Challenge.ID); got.Status != domain.StatusExpired {
		t.Fatalf("read shows expiry: %s", got.Status)
	}
	if _, err := svc.Verify(ctx, second.Challenge.ID, "alice", second.OTP); !errors.Is(err, domain.ErrExpired) {
		t.Fatalf("expired: %v", err)
	}
	if _, err := svc.Verify(ctx, second.Challenge.ID, "alice", "12"); !domain.IsInvalid(err) {
		t.Fatalf("short otp: %v", err)
	}
	if _, err := svc.Initiate(ctx, domain.InitiateRequest{PAN: "7301009999999999", Expiry: "10/31", Amount: 1, Currency: "VND", Merchant: "m"}); !domain.IsConflict(err) {
		t.Fatalf("unknown card: %v", err)
	}
}

func TestConcurrentVerifyCountsEveryAttempt(t *testing.T) {
	svc, _, _ := newService(true)
	ctx := context.Background()
	out, _ := svc.Initiate(ctx, req)
	wrong := "000000"
	if out.OTP == wrong {
		wrong = "111111"
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.Verify(ctx, out.Challenge.ID, "alice", wrong)
		}()
	}
	wg.Wait()
	got, _ := svc.Get(ctx, out.Challenge.ID)
	if got.Status != domain.StatusFailed || got.Attempts != 3 {
		t.Fatalf("attempts must stop at the limit: %+v", got)
	}
}

func TestProviderFailuresDoNotCountAttempts(t *testing.T) {
	svc, _, _, provider := newServiceWithProvider(true)
	ctx := context.Background()
	out, _ := svc.Initiate(ctx, req)
	provider.down = true
	if _, err := svc.Verify(ctx, out.Challenge.ID, "alice", out.OTP); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("authorize-service down: %v", err)
	}
	if got, _ := svc.Get(ctx, out.Challenge.ID); got.Attempts != 0 || got.Status != domain.StatusPending {
		t.Fatalf("an outage must not burn attempts: %+v", got)
	}
	provider.down = false
	if c, err := svc.Verify(ctx, out.Challenge.ID, "alice", out.OTP); err != nil || c.Status != domain.StatusAuthenticated {
		t.Fatalf("verify after recovery: %v", err)
	}

	provider.limit = provider.issued
	if _, err := svc.Initiate(ctx, req); !errors.Is(err, domain.ErrOTPLimit) {
		t.Fatalf("authorize-service otp limit: %v", err)
	}
}
