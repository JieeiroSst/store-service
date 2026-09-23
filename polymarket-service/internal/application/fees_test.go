package application

import (
	"errors"
	"testing"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func feeEnv() *env {
	return newEnvWith(Options{TakerFeeBps: 200, MakerRebateBps: 2000, ReferralBps: 1000})
}

func (e *env) revenue() int64 {
	var sum int64
	for _, x := range e.mem.exch {
		if x.Bucket == model.BucketRevenue {
			sum += x.Amount
		}
	}
	return sum
}

func TestTakerPaysFeeAndItIsSplit(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 100)) // maker: no fee
	res := e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 100))

	// Bob's cash side is 40*100 = 4000; 2% = 80. 20% of it (16) goes to Alice.
	tr := res.Trades[0]
	if tr.TakerFee != 80 || tr.MakerRebate != 16 || tr.ReferralFee != 0 {
		t.Fatalf("fee split %+v", tr)
	}
	if got := e.bal("bob"); got.Available != 10_000-4_000-80 || got.Locked != 0 {
		t.Fatalf("bob %+v", got)
	}
	if got := e.bal("alice"); got.Available != 10_000-6_000+16 {
		t.Fatalf("alice %+v", got)
	}
	if e.revenue() != 64 {
		t.Fatalf("exchange keeps 80-16=64, got %d", e.revenue())
	}
	if res.Order.Fee != 80 {
		t.Fatalf("order fee %d", res.Order.Fee)
	}
	e.checkConservation(t)
}

func TestMakersPayNoFee(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	rest := e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 100))
	if rest.Order.Fee != 0 || e.bal("alice").Locked != 6_000 {
		t.Fatalf("a resting order reserves no fee: locked %d fee %d", e.bal("alice").Locked, rest.Order.Fee)
	}
}

func TestReferrerEarnsCommissionOnRefereeFees(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.mem.referrals["bob"] = model.Referral{RefereeUserID: "bob", ReferrerUserID: "carol"}

	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 100))
	res := e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 100))

	tr := res.Trades[0]
	if tr.ReferralFee != 8 || tr.ReferrerUserID != "carol" || tr.MakerRebate != 16 {
		t.Fatalf("%+v", tr)
	}
	if e.bal("carol").Available != 8 {
		t.Fatalf("carol should earn 8, has %+v", e.bal("carol"))
	}
	if e.revenue() != 80-16-8 {
		t.Fatalf("revenue %d", e.revenue())
	}
	if got, _ := (referralRepo{e.mem}).Earnings(ctx, "carol"); got != 8 {
		t.Fatalf("earnings %d", got)
	}
	// The maker's own referral never matters: only the taker's fee is shared.
	e.mem.referrals["alice"] = model.Referral{RefereeUserID: "alice", ReferrerUserID: "dave"}
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))
	if e.bal("dave").Available != 0 {
		t.Fatal("a maker's referrer earns nothing from the maker's fills")
	}
	e.checkConservation(t)
}

func TestSellFeeComesOutOfProceeds(t *testing.T) {
	e := feeEnv()
	for _, u := range []string{"alice", "bob", "carol"} {
		e.fund(u, 10_000)
	}
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 100))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 100))    // alice holds 100 YES
	e.must(e.limit("carol", model.OutcomeYes, model.Buy, 70, 100)) // resting bid

	before := e.bal("alice").Available
	res := e.must(e.limit("alice", model.OutcomeYes, model.Sell, 70, 100)) // taker sell at 70
	if res.Trades[0].TakerFee != 140 {                                     // 2% of 7000
		t.Fatalf("%+v", res.Trades[0])
	}
	if got := e.bal("alice").Available - before; got != 7_000-140 {
		t.Fatalf("alice should receive proceeds minus fee, got %d", got)
	}
	e.checkConservation(t)
}

func TestFeeReserveIsReleasedWhenOrderRests(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 40))
	res := e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 100)) // 40 filled, 60 rest

	if res.Order.Status != model.OrderOpen || res.Order.Filled != 40 {
		t.Fatalf("%+v", res.Order)
	}
	if got := e.bal("bob"); got.Locked != 40*60 { // only the resting 60 shares at 40... = 2400
		t.Fatalf("resting part must lock exactly price*remaining, locked %d", got.Locked)
	}
	e.checkConservation(t)
}

func TestMarketBuyBudgetIncludesFee(t *testing.T) {
	e := feeEnv()
	for _, u := range []string{"a", "b", "buyer"} {
		e.fund(u, 100_000)
	}
	e.must(e.limit("a", model.OutcomeYes, model.Buy, 50, 200))
	e.must(e.limit("b", model.OutcomeNo, model.Buy, 50, 200))
	e.must(e.limit("a", model.OutcomeYes, model.Sell, 60, 200))

	res, err := e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: "buyer", Outcome: model.OutcomeYes, Side: model.Buy, Type: model.MarketOrder, Amount: 1_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	spent := res.Order.FilledCash + res.Order.Fee
	if spent > 1_000 || res.Order.Filled != 16 { // 16*60=960, fee 19 -> 979; a 17th share (60+1) would need 1040
		t.Fatalf("spent %d shares %d fee %d", spent, res.Order.Filled, res.Order.Fee)
	}
	if got := e.bal("buyer"); got.Available != 100_000-spent || got.Locked != 0 {
		t.Fatalf("unspent budget must return: %+v", got)
	}
	e.checkConservation(t)
}

func TestFailedFillOrKillWithFeesLeavesNoTrace(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	_, err := e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: "bob", Outcome: model.OutcomeNo, Side: model.Buy, Price: 40, Size: 50, TimeInForce: model.FOK,
	})
	if !errors.Is(err, port.ErrNotFilled) {
		t.Fatalf("got %v", err)
	}
	if e.bal("bob").Available != 10_000 || e.revenue() != 0 || len(e.mem.pnl) != 0 {
		t.Fatal("a rolled-back order must not have charged fees")
	}
}

func TestQuoteIncludesFee(t *testing.T) {
	e := feeEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 100))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 100))
	e.must(e.limit("alice", model.OutcomeYes, model.Sell, 60, 100))

	q, err := e.exchange.Quote(ctx, port.QuoteInput{MarketID: e.market.ID, Outcome: model.OutcomeYes, Side: model.Buy, Amount: 612})
	if err != nil {
		t.Fatal(err)
	}
	if q.Shares != 10 || q.Cash != 600 || q.Fee != 12 {
		t.Fatalf("%+v", q)
	}
}

func TestPnLLogTracksRealisedProfitAndFees(t *testing.T) {
	e := feeEnv()
	for _, u := range []string{"alice", "bob", "carol"} {
		e.fund(u, 10_000)
	}
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 100))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 100))
	e.must(e.limit("carol", model.OutcomeYes, model.Buy, 60, 100))
	e.must(e.limit("alice", model.OutcomeYes, model.Sell, 60, 100)) // alice sells 100 at 60, cost 50

	sums := map[string]int64{}
	for _, p := range e.mem.pnl {
		sums[p.UserID] += p.Amount
	}
	// alice: +1000 trade profit, -120 taker fee (2% of 6000), and +20 rebate
	// (20% of bob's 100 fee) from when she was the maker earlier.
	if sums["alice"] != 1_000-120+20 {
		t.Fatalf("alice pnl %d (%+v)", sums["alice"], e.mem.pnl)
	}
}
