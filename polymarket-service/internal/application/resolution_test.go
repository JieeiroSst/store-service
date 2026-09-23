package application

import (
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

// setupPositions leaves alice long 10 YES (cost 600), bob long 10 NO (cost
// 400) and carol with a resting bid.
func (e *env) setupPositions() {
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.fund("carol", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))
	e.must(e.limit("carol", model.OutcomeYes, model.Buy, 20, 5))
}

func TestResolveYesPaysWinnersAndRefundsOpenOrders(t *testing.T) {
	e := newEnv()
	e.setupPositions()

	if _, err := e.resolution.Resolve(ctx, e.market.ID, model.OutcomeYes); err != nil {
		t.Fatal(err)
	}
	if got := e.bal("alice"); got.Available != 9_400+1_000 {
		t.Fatalf("winner should get 100 per share: %+v", got)
	}
	if got := e.bal("bob"); got.Available != 9_600 {
		t.Fatalf("loser gets nothing: %+v", got)
	}
	if got := e.bal("carol"); got.Available != 10_000 || got.Locked != 0 {
		t.Fatalf("resting bid must be refunded: %+v", got)
	}

	alice := e.mem.pos[posKey(e.market.ID, "alice", model.OutcomeYes)]
	bob := e.mem.pos[posKey(e.market.ID, "bob", model.OutcomeNo)]
	if !alice.Settled || alice.Shares != 0 || alice.RealizedPnL != 400 || bob.RealizedPnL != -400 {
		t.Fatalf("alice %+v bob %+v", alice, bob)
	}
	if e.mem.events[e.market.EventID].Status != model.EventResolved {
		t.Fatal("event with all markets resolved must be resolved")
	}
	// Everything ever deposited is now cash: no supply remains.
	var cash int64
	for _, b := range e.mem.balances {
		cash += b.Available + b.Locked
	}
	for _, x := range e.mem.exch {
		cash += x.Amount
	}
	if cash != e.deposited {
		t.Fatalf("cash %d != deposited %d", cash, e.deposited)
	}
	if _, err := e.resolution.Resolve(ctx, e.market.ID, model.OutcomeNo); !errors.Is(err, port.ErrInvalidTransition) {
		t.Fatalf("a resolved market cannot be resolved again, got %v", err)
	}
}

func TestResolveSplitPaysHalfBothSides(t *testing.T) {
	e := newEnv()
	e.setupPositions()
	if _, err := e.resolution.Resolve(ctx, e.market.ID, model.OutcomeSplit); err != nil {
		t.Fatal(err)
	}
	if a, b := e.bal("alice").Available, e.bal("bob").Available; a != 9_400+500 || b != 9_600+500 {
		t.Fatalf("alice %d bob %d", a, b)
	}
}

func TestProposeDisputeFinalizeFlow(t *testing.T) {
	e := newEnv()
	e.setupPositions()
	now := time.Now()
	e.resolution.now = func() time.Time { return now }
	e.exchange.now = func() time.Time { return now }

	m, err := e.resolution.Propose(ctx, e.market.ID, model.OutcomeYes)
	if err != nil || m.Status != model.MarketProposed {
		t.Fatalf("propose: %+v %v", m, err)
	}
	if _, err := e.limit("alice", model.OutcomeYes, model.Buy, 50, 1); !errors.Is(err, port.ErrMarketNotTradable) {
		t.Fatalf("trading must halt during the dispute window, got %v", err)
	}
	if _, err := e.resolution.Finalize(ctx, e.market.ID); !errors.Is(err, port.ErrDisputeWindowOpen) {
		t.Fatalf("finalize inside the window: %v", err)
	}

	if _, err := e.resolution.Dispute(ctx, e.market.ID, "bob", "wrong source"); err != nil {
		t.Fatal(err)
	}
	e.resolution.now = func() time.Time { return now.Add(3 * time.Hour) }
	if _, err := e.resolution.Finalize(ctx, e.market.ID); !errors.Is(err, port.ErrInvalidTransition) {
		t.Fatalf("a disputed market needs an admin ruling, got %v", err)
	}
	if m, err := e.resolution.Resolve(ctx, e.market.ID, model.OutcomeNo); err != nil || m.ResolvedOutcome != model.OutcomeNo {
		t.Fatalf("admin ruling: %+v %v", m, err)
	}
	if got := e.bal("bob").Available; got != 9_600+1_000 {
		t.Fatalf("bob wins after the ruling: %d", got)
	}
}

func TestFinalizeAfterWindowSettlesProposal(t *testing.T) {
	e := newEnv()
	e.setupPositions()
	now := time.Now()
	e.resolution.now = func() time.Time { return now }
	if _, err := e.resolution.Propose(ctx, e.market.ID, model.OutcomeYes); err != nil {
		t.Fatal(err)
	}

	e.resolution.now = func() time.Time { return now.Add(3 * time.Hour) }
	if _, err := e.resolution.Dispute(ctx, e.market.ID, "bob", "late"); !errors.Is(err, port.ErrDisputeWindowShut) {
		t.Fatalf("late dispute: %v", err)
	}
	m, err := e.resolution.Finalize(ctx, e.market.ID)
	if err != nil || m.ResolvedOutcome != model.OutcomeYes {
		t.Fatalf("finalize: %+v %v", m, err)
	}
	if got := e.bal("alice").Available; got != 9_400+1_000 {
		t.Fatalf("alice paid %d", got)
	}
}
