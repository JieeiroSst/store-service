package application

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func TestSummaryIssuesACodeOnceAndReportsEarnings(t *testing.T) {
	e := feeEnv()
	first, err := e.referral.Summary(ctx, "carol")
	if err != nil || first.RefCode != "CODE123" || first.CommissionBps != 1000 {
		t.Fatalf("%+v %v", first, err)
	}
	e.referrals.err = port.ErrUpstream // a second call must not need referral-service at all
	second, err := e.referral.Summary(ctx, "carol")
	if err != nil || second.RefCode != first.RefCode {
		t.Fatalf("the code must be reused: %+v %v", second, err)
	}
	e.referrals.err = nil

	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.referrals.owner = "carol"
	if _, err := e.referral.Redeem(ctx, "bob", "CODE123"); err != nil {
		t.Fatal(err)
	}
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 100))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 100))

	s, err := e.referral.Summary(ctx, "carol")
	if err != nil || s.Referees != 1 || s.Earnings == 0 || s.ServiceStats == nil {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestRedeemAttributesANewUserOnce(t *testing.T) {
	e := feeEnv()
	e.referrals.owner = "carol"
	r, err := e.referral.Redeem(ctx, "bob", "CODE123")
	if err != nil || r.ReferrerUserID != "carol" || r.RefereeUserID != "bob" {
		t.Fatalf("%+v %v", r, err)
	}
	if got := e.referrals.activate; len(got) != 1 || got[0] != "CODE123:bob" {
		t.Fatalf("referral-service must be told: %v", got)
	}
	if _, err := e.referral.Redeem(ctx, "bob", "OTHER"); !errors.Is(err, port.ErrReferralNotAllowed) {
		t.Fatalf("second redeem: %v", err)
	}
	if len(e.notifier.sent) == 0 || e.notifier.sent[len(e.notifier.sent)-1].UserID != "carol" {
		t.Fatal("the referrer should be notified")
	}
}

func TestRedeemIsForNewUsersOnly(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))

	if _, err := e.referral.Redeem(ctx, "bob", "CODE123"); !errors.Is(err, port.ErrReferralNotAllowed) {
		t.Fatalf("a user who already traded cannot claim a referrer: %v", err)
	}
	if len(e.referrals.activate) != 0 {
		t.Fatal("referral-service must not be called for an ineligible user")
	}
}

func TestRedeemRejectsSelfReferralAndBadCodes(t *testing.T) {
	e := feeEnv()
	e.referrals.owner = "bob"
	if _, err := e.referral.Redeem(ctx, "bob", "MINE"); !errors.Is(err, port.ErrReferralNotAllowed) {
		t.Fatalf("self referral: %v", err)
	}
	e.referrals.owner, e.referrals.err = "", port.ErrInvalidInput
	if _, err := e.referral.Redeem(ctx, "dave", "NOPE"); !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("unknown code: %v", err)
	}
	if _, err := e.referral.Redeem(ctx, "dave", ""); !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("empty code: %v", err)
	}
	if _, ok := e.mem.referrals["dave"]; ok {
		t.Fatal("nothing may be stored for a failed redeem")
	}
}

type countingDirectory struct {
	calls atomic.Int64
	ids   map[string]string
}

func (d *countingDirectory) Validate(_ context.Context, token string) (string, error) {
	d.calls.Add(1)
	if id, ok := d.ids[token]; ok {
		return id, nil
	}
	return "", port.ErrUnauthenticated
}

func TestIdentityCachesSuccessesButNeverFailures(t *testing.T) {
	dir := &countingDirectory{ids: map[string]string{"good": "42"}}
	s := NewIdentityService(dir)
	now := time.Now()
	s.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if id, err := s.Authenticate(ctx, "good"); err != nil || id != "42" {
			t.Fatalf("%q %v", id, err)
		}
	}
	if dir.calls.Load() != 1 {
		t.Fatalf("a valid token is looked up once per TTL, not per request: %d calls", dir.calls.Load())
	}

	now = now.Add(time.Minute)
	if _, err := s.Authenticate(ctx, "good"); err != nil || dir.calls.Load() != 2 {
		t.Fatalf("expired entries are looked up again: %d", dir.calls.Load())
	}

	for i := 0; i < 2; i++ {
		if _, err := s.Authenticate(ctx, "bad"); !errors.Is(err, port.ErrUnauthenticated) {
			t.Fatalf("%v", err)
		}
	}
	if dir.calls.Load() != 4 {
		t.Fatalf("failures must not be cached: %d calls", dir.calls.Load())
	}
	if _, err := s.Authenticate(ctx, ""); !errors.Is(err, port.ErrUnauthenticated) || dir.calls.Load() != 4 {
		t.Fatalf("an empty token is refused without a lookup")
	}
}
