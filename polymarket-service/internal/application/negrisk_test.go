package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

// negRiskEnv adds a three-way neg-risk event (Alice, Bob, Carol run for mayor)
// to a fresh environment.
func negRiskEnv(t *testing.T, opts Options) (*env, *model.Event) {
	t.Helper()
	e := newEnvWith(opts)
	catalog := NewCatalogService(eventRepo{e.mem}, e.mem, e.mem, Options{ShareValue: 100})
	ev, err := catalog.CreateEvent(ctx, port.CreateEventInput{
		Title: "Who wins?", NegRisk: true, EndDate: time.Now().Add(time.Hour),
		Markets: []port.CreateMarketInput{
			{Question: "Alice?", GroupItemTitle: "Alice"}, {Question: "Bob?", GroupItemTitle: "Bob"}, {Question: "Carol?", GroupItemTitle: "Carol"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return e, ev
}

func (e *env) on(marketID int64, user string, o model.Outcome, s model.Side, price, size int64) {
	_, err := e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{MarketID: marketID, UserID: user, Outcome: o, Side: s, Price: price, Size: size})
	if err != nil {
		panic(err)
	}
}

// holdNo gives user shares NO in a market by minting against a counterparty.
func (e *env) holdNo(marketID int64, user, counterparty string, price, size int64) {
	e.on(marketID, counterparty, model.OutcomeYes, model.Buy, 100-price, size)
	e.on(marketID, user, model.OutcomeNo, model.Buy, price, size)
}

func (e *env) sharesIn(marketID int64, user string, o model.Outcome) int64 {
	return e.mem.pos[posKey(marketID, user, o)].Shares
}

func TestConvertOneNoIntoYesInEveryOtherMarket(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	a, b, c := ev.Markets[0].ID, ev.Markets[1].ID, ev.Markets[2].ID
	e.fund("alice", 100_000)
	e.fund("bob", 100_000)
	e.holdNo(a, "alice", "bob", 60, 10) // alice: 10 NO in market A, cost 600

	res, err := e.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a}, Amount: 10})
	if err != nil {
		t.Fatal(err)
	}
	if res.Cash != 0 || len(res.YesMarkets) != 2 {
		t.Fatalf("%+v", res)
	}
	if e.sharesIn(a, "alice", model.OutcomeNo) != 0 || e.sharesIn(b, "alice", model.OutcomeYes) != 10 || e.sharesIn(c, "alice", model.OutcomeYes) != 10 {
		t.Fatal("NO in A must become YES in B and C")
	}
	if cost := e.mem.pos[posKey(b, "alice", model.OutcomeYes)].CostBasis + e.mem.pos[posKey(c, "alice", model.OutcomeYes)].CostBasis; cost != 600 {
		t.Fatalf("cost basis must carry over, got %d", cost)
	}
}

func TestConvertSeveralNosPaysCash(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	a, b, c := ev.Markets[0].ID, ev.Markets[1].ID, ev.Markets[2].ID
	e.fund("alice", 100_000)
	e.fund("bob", 100_000)
	e.holdNo(a, "alice", "bob", 60, 10)
	e.holdNo(b, "alice", "bob", 70, 10)
	before := e.bal("alice").Available

	res, err := e.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a, b}, Amount: 10})
	if err != nil {
		t.Fatal(err)
	}
	// 2 NOs -> 1 YES (in C) + 1 share value of cash per share converted.
	if res.Cash != 10*100 || len(res.YesMarkets) != 1 || res.YesMarkets[0] != c {
		t.Fatalf("%+v", res)
	}
	if e.bal("alice").Available-before != 1_000 {
		t.Fatal("cash not credited")
	}
	// The NOs cost 600+700=1300, cash returned 1000, so the YES carries 300.
	if got := e.mem.pos[posKey(c, "alice", model.OutcomeYes)].CostBasis; got != 300 {
		t.Fatalf("YES cost basis %d", got)
	}
}

func TestConvertRules(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	a := ev.Markets[0].ID
	e.fund("alice", 100_000)
	e.fund("bob", 100_000)
	e.holdNo(a, "alice", "bob", 60, 10)

	cases := map[string]struct {
		in   port.ConvertInput
		want error
	}{
		"more than held": {port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a}, Amount: 11}, port.ErrInsufficientShares},
		"foreign market": {port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{e.market.ID}, Amount: 1}, port.ErrInvalidInput},
		"no markets":     {port.ConvertInput{EventID: ev.ID, UserID: "alice", Amount: 1}, port.ErrInvalidInput},
		"zero amount":    {port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a}}, port.ErrInvalidSize},
		"not a neg-risk": {port.ConvertInput{EventID: e.market.EventID, UserID: "alice", MarketIDs: []int64{e.market.ID}, Amount: 1}, port.ErrNotNegRisk},
		"holds nothing":  {port.ConvertInput{EventID: ev.ID, UserID: "bob", MarketIDs: []int64{a}, Amount: 1}, port.ErrInsufficientShares},
	}
	for name, c := range cases {
		if _, err := e.exchange.Convert(ctx, c.in); !errors.Is(err, c.want) {
			t.Errorf("%s: want %v, got %v", name, c.want, err)
		}
	}
}

func TestConvertCannotUseSharesLockedInOrders(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	a := ev.Markets[0].ID
	e.fund("alice", 100_000)
	e.fund("bob", 100_000)
	e.holdNo(a, "alice", "bob", 60, 10)
	e.on(a, "alice", model.OutcomeNo, model.Sell, 90, 10) // locks all 10 NO

	if _, err := e.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a}, Amount: 1}); !errors.Is(err, port.ErrInsufficientShares) {
		t.Fatalf("got %v", err)
	}
}

// After any conversion the event must still pay out exactly its collateral
// whichever market wins. Deposits are the only money in, so cash after
// resolution must equal deposits.
func TestConversionKeepsEventSolventForEveryWinner(t *testing.T) {
	for winner := 0; winner < 3; winner++ {
		e, ev := negRiskEnv(t, Options{TakerFeeBps: 100, MakerRebateBps: 2000})
		a, b, c := ev.Markets[0].ID, ev.Markets[1].ID, ev.Markets[2].ID
		for _, u := range []string{"alice", "bob", "carol"} {
			e.fund(u, 1_000_000)
		}
		e.holdNo(a, "alice", "bob", 60, 10)
		e.holdNo(b, "alice", "carol", 70, 10)
		e.holdNo(c, "alice", "bob", 80, 10)
		e.holdNo(a, "carol", "bob", 55, 7)

		if _, err := e.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a, b}, Amount: 6}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{c}, Amount: 4}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "carol", MarketIDs: []int64{a}, Amount: 7}); err != nil {
			t.Fatal(err)
		}

		if _, err := e.resolution.ResolveEvent(ctx, ev.ID, ev.Markets[winner].ID); err != nil {
			t.Fatal(err)
		}
		var total int64
		for _, bal := range e.mem.balances {
			if bal.Locked != 0 {
				t.Fatalf("winner %d: locked cash left over", winner)
			}
			total += bal.Available
		}
		for _, x := range e.mem.exch {
			total += x.Amount
		}
		if total != e.deposited {
			t.Fatalf("winner %d: after resolution %d != deposited %d", winner, total, e.deposited)
		}
	}
}

func TestNegRiskEventsResolveAsAWhole(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	a, b := ev.Markets[0].ID, ev.Markets[1].ID

	if _, err := e.resolution.Resolve(ctx, a, model.OutcomeYes); !errors.Is(err, port.ErrNegRisk) {
		t.Fatalf("market-level resolve: %v", err)
	}
	if _, err := e.resolution.Propose(ctx, a, model.OutcomeYes); !errors.Is(err, port.ErrNegRisk) {
		t.Fatalf("market-level propose: %v", err)
	}
	if _, err := e.resolution.ProposeEvent(ctx, e.market.EventID, a); !errors.Is(err, port.ErrNotNegRisk) {
		t.Fatalf("event endpoints are for neg-risk events only: %v", err)
	}
	if _, err := e.resolution.ResolveEvent(ctx, ev.ID, e.market.ID); !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("winner must belong to the event: %v", err)
	}

	if _, err := e.resolution.ResolveEvent(ctx, ev.ID, b); err != nil {
		t.Fatal(err)
	}
	for i, m := range ev.Markets {
		got := e.mem.markets[m.ID]
		want := model.OutcomeNo
		if m.ID == b {
			want = model.OutcomeYes
		}
		if got.Status != model.MarketResolved || got.ResolvedOutcome != want {
			t.Fatalf("market %d: %+v", i, got)
		}
	}
	if e.mem.events[ev.ID].Status != model.EventResolved {
		t.Fatal("event must be resolved")
	}
}

func TestNegRiskProposalDisputeAndFinalizeAffectTheWholeEvent(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	a, b := ev.Markets[0].ID, ev.Markets[1].ID
	now := time.Now()
	e.resolution.now = func() time.Time { return now }
	e.exchange.now = func() time.Time { return now }

	if _, err := e.resolution.ProposeEvent(ctx, ev.ID, b); err != nil {
		t.Fatal(err)
	}
	for _, m := range ev.Markets {
		got := e.mem.markets[m.ID]
		if got.Status != model.MarketProposed || got.DisputeDeadline == nil {
			t.Fatalf("every market must enter the window: %+v", got)
		}
	}
	if e.mem.markets[b].ProposedOutcome != model.OutcomeYes || e.mem.markets[a].ProposedOutcome != model.OutcomeNo {
		t.Fatal("winner proposed YES, the rest NO")
	}

	// A dispute on one market freezes all of them.
	if _, err := e.resolution.Dispute(ctx, a, "bob", "wrong"); err != nil {
		t.Fatal(err)
	}
	for _, m := range ev.Markets {
		if e.mem.markets[m.ID].Status != model.MarketDisputed {
			t.Fatalf("market %d not disputed", m.ID)
		}
	}
	e.resolution.now = func() time.Time { return now.Add(3 * time.Hour) }
	if _, err := e.resolution.Finalize(ctx, a); !errors.Is(err, port.ErrInvalidTransition) {
		t.Fatalf("a disputed event needs a ruling: %v", err)
	}
	if _, err := e.resolution.ResolveEvent(ctx, ev.ID, ev.Markets[2].ID); err != nil {
		t.Fatal(err)
	}
}

func TestNegRiskFinalizeSettlesEveryMarket(t *testing.T) {
	e, ev := negRiskEnv(t, Options{})
	now := time.Now()
	e.resolution.now = func() time.Time { return now }
	if _, err := e.resolution.ProposeEvent(ctx, ev.ID, ev.Markets[1].ID); err != nil {
		t.Fatal(err)
	}
	e.resolution.now = func() time.Time { return now.Add(3 * time.Hour) }
	if _, err := e.resolution.Finalize(ctx, ev.Markets[2].ID); err != nil {
		t.Fatal(err)
	}
	for _, m := range ev.Markets {
		if e.mem.markets[m.ID].Status != model.MarketResolved {
			t.Fatalf("market %d not resolved", m.ID)
		}
	}
}

var _ = context.Background
