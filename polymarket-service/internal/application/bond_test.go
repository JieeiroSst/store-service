package application

import (
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func bondEnv(t *testing.T) *env {
	t.Helper()
	e := newEnvWith(Options{DisputeBond: 500})
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.fund("carol", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))
	now := time.Now()
	e.resolution.now = func() time.Time { return now }
	if _, err := e.resolution.Propose(ctx, e.market.ID, model.OutcomeYes); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e *env) bonds() int64 {
	var sum int64
	for _, x := range e.mem.exch {
		if x.Bucket == model.BucketBond {
			sum += x.Amount
		}
	}
	return sum
}

func TestDisputeTakesABond(t *testing.T) {
	e := bondEnv(t)
	if _, err := e.resolution.Dispute(ctx, e.market.ID, "bob", "wrong"); err != nil {
		t.Fatal(err)
	}
	if got := e.bal("bob").Available; got != 9_600-500 {
		t.Fatalf("bob should have posted 500, has %d", got)
	}
	if e.bonds() != 500 || e.mem.markets[e.market.ID].DisputeBond != 500 {
		t.Fatalf("bond not recorded: held %d market %+v", e.bonds(), e.mem.markets[e.market.ID])
	}
}

func TestDisputeNeedsTheFunds(t *testing.T) {
	e := bondEnv(t)
	if _, err := e.resolution.Dispute(ctx, e.market.ID, "nobody", "x"); !errors.Is(err, port.ErrInsufficientBalance) {
		t.Fatalf("got %v", err)
	}
	if e.mem.markets[e.market.ID].Status != model.MarketProposed {
		t.Fatal("a refused dispute must not change the market")
	}
}

func TestBondIsRefundedWhenTheRulingOverturnsTheProposal(t *testing.T) {
	e := bondEnv(t)
	if _, err := e.resolution.Dispute(ctx, e.market.ID, "bob", "wrong"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.resolution.Resolve(ctx, e.market.ID, model.OutcomeNo); err != nil { // proposal was YES
		t.Fatal(err)
	}
	// bob won the market (NO) and got his bond back
	if got := e.bal("bob").Available; got != 9_600+1_000 {
		t.Fatalf("bob %d", got)
	}
	if e.bonds() != 0 || e.revenue() != 0 {
		t.Fatalf("bond must be released without touching revenue: bonds %d revenue %d", e.bonds(), e.revenue())
	}
	if !e.mem.markets[e.market.ID].BondSettled {
		t.Fatal("bond must be marked settled")
	}
}

func TestBondIsForfeitedWhenTheProposalStands(t *testing.T) {
	e := bondEnv(t)
	if _, err := e.resolution.Dispute(ctx, e.market.ID, "carol", "spite"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.resolution.Resolve(ctx, e.market.ID, model.OutcomeYes); err != nil { // same as proposed
		t.Fatal(err)
	}
	if got := e.bal("carol").Available; got != 10_000-500 {
		t.Fatalf("carol loses her bond, has %d", got)
	}
	if e.bonds() != 0 || e.revenue() != 500 {
		t.Fatalf("forfeited bond becomes revenue: bonds %d revenue %d", e.bonds(), e.revenue())
	}
	var lost int64
	for _, p := range e.mem.pnl {
		if p.UserID == "carol" && p.Kind == model.PnLBond {
			lost += p.Amount
		}
	}
	if lost != -500 {
		t.Fatalf("the loss must show in carol's pnl, got %d", lost)
	}
	e.checkConservation(t)
}

func TestNoBondWhenConfiguredOff(t *testing.T) {
	e := newEnv()
	e.fund("alice", 100)
	e.resolution.now = func() time.Time { return time.Now() }
	if _, err := e.resolution.Propose(ctx, e.market.ID, model.OutcomeYes); err != nil {
		t.Fatal(err)
	}
	if _, err := e.resolution.Dispute(ctx, e.market.ID, "anyone", "x"); err != nil {
		t.Fatalf("with no bond configured anyone may dispute: %v", err)
	}
}
