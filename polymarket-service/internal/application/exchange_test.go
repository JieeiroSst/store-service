package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

var ctx = context.Background()

func (e *env) limit(user string, o model.Outcome, s model.Side, price, size int64) (*port.PlaceOrderResult, error) {
	return e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: user, Outcome: o, Side: s, Price: price, Size: size,
	})
}

func (e *env) must(res *port.PlaceOrderResult, err error) *port.PlaceOrderResult {
	if err != nil {
		panic(err)
	}
	return res
}

func (e *env) bal(user string) model.Balance { return e.mem.balances[user] }

func (e *env) shares(user string, o model.Outcome) int64 {
	return e.mem.pos[posKey(e.market.ID, user, o)].Shares
}

// checkConservation asserts the exchange is solvent: every unit of cash that
// was ever deposited is either in a balance or backing an outstanding
// YES/NO pair worth one share value.
func (e *env) checkConservation(t *testing.T) {
	t.Helper()
	var cash, yes, no, lockedShares int64
	for _, b := range e.mem.balances {
		if b.Available < 0 || b.Locked < 0 {
			t.Fatalf("negative balance %+v", b)
		}
		cash += b.Available + b.Locked
	}
	var lockedCash int64
	for _, o := range e.mem.orders {
		if o.Status == model.OrderOpen && o.Side == model.Buy {
			lockedCash += o.LockedCash + o.FeeReserve
		}
	}
	for _, b := range e.mem.balances {
		lockedCash -= b.Locked
	}
	if lockedCash != 0 {
		t.Fatalf("balance.Locked disagrees with open buy orders by %d", lockedCash)
	}
	for _, p := range e.mem.pos {
		lockedShares += p.LockedShares
		if p.Outcome == model.OutcomeYes {
			yes += p.Shares
		} else {
			no += p.Shares
		}
	}
	if yes != no {
		t.Fatalf("YES supply %d != NO supply %d", yes, no)
	}
	var openSells int64
	for _, o := range e.mem.orders {
		if o.Status == model.OrderOpen && o.Side == model.Sell {
			openSells += o.Remaining()
		}
	}
	if openSells != lockedShares {
		t.Fatalf("locked shares %d != open sell size %d", lockedShares, openSells)
	}
	var exchange int64
	for _, x := range e.mem.exch {
		exchange += x.Amount
	}
	if got := cash + exchange + yes*e.market.ShareValue; got != e.deposited {
		t.Fatalf("cash %d + exchange %d + backing %d = %d, deposited %d", cash, exchange, yes*e.market.ShareValue, got, e.deposited)
	}
}

func TestBuyYesMeetsBuyNoAndMints(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)

	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	if e.bal("alice").Locked != 600 {
		t.Fatalf("resting buy should lock 600, got %d", e.bal("alice").Locked)
	}
	res := e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))

	if len(res.Trades) != 1 || res.Trades[0].Kind != model.KindMint || res.Trades[0].YesPrice != 60 {
		t.Fatalf("expected one mint at 60, got %+v", res.Trades)
	}
	if e.shares("alice", model.OutcomeYes) != 10 || e.shares("bob", model.OutcomeNo) != 10 {
		t.Fatal("shares not minted to both buyers")
	}
	if a, b := e.bal("alice"), e.bal("bob"); a.Available != 9_400 || a.Locked != 0 || b.Available != 9_600 {
		t.Fatalf("balances alice=%+v bob=%+v", a, b)
	}
	if res.Order.Status != model.OrderFilled {
		t.Fatalf("taker should be filled, got %s", res.Order.Status)
	}
	e.checkConservation(t)
}

func TestTakerGetsPriceImprovement(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)

	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 70, 10))
	// Bob would pay up to 40 for NO (yes-price 60) but trades at Alice's 70,
	// i.e. only 30 per NO share.
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))

	if got := e.bal("bob"); got.Available != 9_700 || got.Locked != 0 {
		t.Fatalf("bob should have paid 300 in total, got %+v", got)
	}
	if got := e.bal("alice"); got.Available != 9_300 {
		t.Fatalf("alice should have paid 700, got %+v", got)
	}
	e.checkConservation(t)
}

func TestSwapSellYesToBuyYes(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.fund("carol", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 10)) // alice holds 10 YES at cost 500

	e.must(e.limit("alice", model.OutcomeYes, model.Sell, 55, 10))
	if got := e.mem.pos[posKey(e.market.ID, "alice", model.OutcomeYes)]; got.LockedShares != 10 {
		t.Fatalf("resting sell must lock shares, got %+v", got)
	}
	if _, err := e.limit("alice", model.OutcomeYes, model.Sell, 56, 1); !errors.Is(err, port.ErrInsufficientShares) {
		t.Fatalf("locked shares must not be sellable twice, got %v", err)
	}

	res := e.must(e.limit("carol", model.OutcomeYes, model.Buy, 60, 10))
	if len(res.Trades) != 1 || res.Trades[0].Kind != model.KindSwap || res.Trades[0].YesPrice != 55 {
		t.Fatalf("expected one swap at the maker's 55, got %+v", res.Trades)
	}
	alice := e.mem.pos[posKey(e.market.ID, "alice", model.OutcomeYes)]
	if alice.Shares != 0 || alice.RealizedPnL != 50 { // sold 550 against a 500 cost
		t.Fatalf("alice position %+v", alice)
	}
	if e.shares("carol", model.OutcomeYes) != 10 || e.bal("carol").Available != 9_450 {
		t.Fatalf("carol %+v shares %d", e.bal("carol"), e.shares("carol", model.OutcomeYes))
	}
	e.checkConservation(t)
}

func TestMergeReleasesCollateral(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 10))

	e.must(e.limit("alice", model.OutcomeYes, model.Sell, 40, 10))     // ask 40
	res := e.must(e.limit("bob", model.OutcomeNo, model.Sell, 55, 10)) // bid 45 crosses ask 40

	if len(res.Trades) != 1 || res.Trades[0].Kind != model.KindMerge || res.Trades[0].YesPrice != 40 {
		t.Fatalf("expected a merge at the maker's 40, got %+v", res.Trades)
	}
	if e.shares("alice", model.OutcomeYes) != 0 || e.shares("bob", model.OutcomeNo) != 0 {
		t.Fatal("merge must burn both sides")
	}
	// Alice gets 40 per share, Bob 60: together exactly the 100 released.
	if a, b := e.bal("alice").Available, e.bal("bob").Available; a != 9_500+400 || b != 9_500+600 {
		t.Fatalf("alice=%d bob=%d", a, b)
	}
	e.checkConservation(t)
}

func TestPriceTimePriority(t *testing.T) {
	e := newEnv()
	for _, u := range []string{"a", "b", "c", "taker"} {
		e.fund(u, 100_000)
	}
	// Two NO buyers create YES inventory for the sellers.
	e.must(e.limit("a", model.OutcomeYes, model.Buy, 50, 5))
	e.must(e.limit("b", model.OutcomeYes, model.Buy, 50, 5))
	e.must(e.limit("c", model.OutcomeNo, model.Buy, 50, 10))

	e.must(e.limit("a", model.OutcomeYes, model.Sell, 60, 5)) // first at 60
	e.must(e.limit("b", model.OutcomeYes, model.Sell, 60, 5)) // second at 60

	res := e.must(e.limit("taker", model.OutcomeYes, model.Buy, 60, 7))
	if len(res.Trades) != 2 || res.Trades[0].MakerUserID != "a" || res.Trades[0].Size != 5 || res.Trades[1].MakerUserID != "b" || res.Trades[1].Size != 2 {
		t.Fatalf("time priority violated: %+v", res.Trades)
	}
	book, _ := e.exchange.OrderBook(ctx, e.market.ID, model.OutcomeYes)
	if len(book.Asks) != 1 || book.Asks[0].Size != 3 {
		t.Fatalf("book should keep 3 shares of b's ask, got %+v", book.Asks)
	}
	e.checkConservation(t)
}

func TestPartialFillRestsRemainder(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 4))

	res := e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 10))
	if res.Order.Status != model.OrderOpen || res.Order.Filled != 4 {
		t.Fatalf("expected 4 filled and resting, got %+v", res.Order)
	}
	if got := e.bal("bob"); got.Locked != 300 || got.Available != 10_000-200-300 {
		t.Fatalf("bob %+v", got)
	}
	book, _ := e.exchange.OrderBook(ctx, e.market.ID, model.OutcomeNo)
	if len(book.Bids) != 1 || book.Bids[0].YesPrice != 50 || book.Bids[0].Size != 6 {
		t.Fatalf("NO book should show a 6-share bid at 50, got %+v", book.Bids)
	}
	e.checkConservation(t)
}

func TestFillOrKillRollsBack(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 4))

	_, err := e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: "bob", Outcome: model.OutcomeNo, Side: model.Buy,
		Price: 50, Size: 10, TimeInForce: model.FOK,
	})
	if !errors.Is(err, port.ErrNotFilled) {
		t.Fatalf("want ErrNotFilled, got %v", err)
	}
	if e.bal("bob").Available != 10_000 || e.bal("alice").Locked != 200 || len(e.mem.trades) != 0 {
		t.Fatal("a failed FOK must leave no trace")
	}
	e.checkConservation(t)
}

func TestFillAndKillCancelsRemainder(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 4))

	res, err := e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: "bob", Outcome: model.OutcomeNo, Side: model.Buy,
		Price: 50, Size: 10, TimeInForce: model.FAK,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Order.Status != model.OrderCanceled || res.Order.Filled != 4 {
		t.Fatalf("got %+v", res.Order)
	}
	if got := e.bal("bob"); got.Locked != 0 || got.Available != 10_000-200 {
		t.Fatalf("remainder must be refunded, got %+v", got)
	}
	e.checkConservation(t)
}

func TestMarketBuySpendsBudgetAcrossLevels(t *testing.T) {
	e := newEnv()
	for _, u := range []string{"a", "b", "buyer"} {
		e.fund(u, 100_000)
	}
	e.must(e.limit("a", model.OutcomeYes, model.Buy, 50, 20))
	e.must(e.limit("b", model.OutcomeNo, model.Buy, 50, 20))
	e.must(e.limit("a", model.OutcomeYes, model.Sell, 60, 5))
	e.must(e.limit("a", model.OutcomeYes, model.Sell, 70, 5))

	q, err := e.exchange.Quote(ctx, port.QuoteInput{MarketID: e.market.ID, Outcome: model.OutcomeYes, Side: model.Buy, Amount: 500})
	if err != nil || q.Shares != 5+2 || q.Cash != 5*60+2*70 {
		t.Fatalf("quote %+v err %v", q, err)
	}

	res, err := e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: "buyer", Outcome: model.OutcomeYes, Side: model.Buy, Type: model.MarketOrder, Amount: 500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Order.Filled != 7 || res.Order.FilledCash != 440 || res.Order.Status != model.OrderFilled {
		t.Fatalf("got %+v", res.Order)
	}
	if got := e.bal("buyer"); got.Available != 100_000-440 || got.Locked != 0 {
		t.Fatalf("unspent budget must be refunded, got %+v", got)
	}
	e.checkConservation(t)
}

func TestCancelReturnsEscrow(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	res := e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 10))

	if _, err := e.exchange.CancelOrder(ctx, res.Order.ID, "mallory"); !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("only the owner may cancel, got %v", err)
	}
	if _, err := e.exchange.CancelOrder(ctx, res.Order.ID, "alice"); err != nil {
		t.Fatal(err)
	}
	if got := e.bal("alice"); got.Available != 10_000 || got.Locked != 0 {
		t.Fatalf("got %+v", got)
	}
	if _, err := e.exchange.CancelOrder(ctx, res.Order.ID, "alice"); !errors.Is(err, port.ErrOrderNotOpen) {
		t.Fatalf("second cancel: %v", err)
	}
	e.checkConservation(t)
}

func TestSelfTradeCancelsRestingOrder(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	resting := e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 10))

	res := e.must(e.limit("alice", model.OutcomeNo, model.Buy, 50, 10))
	if len(res.Trades) != 0 {
		t.Fatal("a user must never trade with themselves")
	}
	if got := e.mem.orders[resting.Order.ID]; got.Status != model.OrderCanceled {
		t.Fatalf("resting order should be cancelled, got %s", got.Status)
	}
	e.checkConservation(t)
}

func TestExpiredMakerIsSkipped(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	soon := e.exchange.now().Add(time.Minute)
	e.must(e.exchange.PlaceOrder(ctx, port.PlaceOrderInput{
		MarketID: e.market.ID, UserID: "alice", Outcome: model.OutcomeYes, Side: model.Buy, Price: 50, Size: 10, ExpiresAt: &soon,
	}))
	e.exchange.now = func() time.Time { return soon.Add(time.Second) }

	res := e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 10))
	if len(res.Trades) != 0 {
		t.Fatal("expired orders must not trade")
	}
	if got := e.bal("alice"); got.Locked != 0 || got.Available != 10_000 {
		t.Fatalf("expired order's escrow must return, got %+v", got)
	}
	e.checkConservation(t)
}

func TestOrderValidation(t *testing.T) {
	e := newEnv()
	e.fund("alice", 100)
	cases := []struct {
		name  string
		price int64
		size  int64
		want  error
	}{
		{"price zero", 0, 5, port.ErrInvalidPrice},
		{"price at share value", 100, 5, port.ErrInvalidPrice},
		{"zero size", 50, 0, port.ErrInvalidSize},
		{"too expensive", 50, 5, port.ErrInsufficientBalance},
	}
	for _, c := range cases {
		if _, err := e.limit("alice", model.OutcomeYes, model.Buy, c.price, c.size); !errors.Is(err, c.want) {
			t.Errorf("%s: want %v, got %v", c.name, c.want, err)
		}
	}
	if _, err := e.limit("alice", model.OutcomeYes, model.Sell, 50, 1); !errors.Is(err, port.ErrInsufficientShares) {
		t.Errorf("selling nothing: got %v", err)
	}
}

func TestClientOrderIDIsIdempotent(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	in := port.PlaceOrderInput{MarketID: e.market.ID, UserID: "alice", ClientOrderID: "abc", Outcome: model.OutcomeYes, Side: model.Buy, Price: 50, Size: 10}
	first := e.must(e.exchange.PlaceOrder(ctx, in))
	second := e.must(e.exchange.PlaceOrder(ctx, in))
	if first.Order.ID != second.Order.ID || e.bal("alice").Locked != 500 {
		t.Fatal("retry with the same client_order_id must not place a second order")
	}
}

func TestClosedMarketRejectsOrders(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.exchange.now = func() time.Time { return e.market.EndTime.Add(time.Second) }
	if _, err := e.limit("alice", model.OutcomeYes, model.Buy, 50, 1); !errors.Is(err, port.ErrMarketNotTradable) {
		t.Fatalf("got %v", err)
	}
}

func TestOrderBookNoViewMirrorsYes(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 30, 10))

	yes, _ := e.exchange.OrderBook(ctx, e.market.ID, model.OutcomeYes)
	no, _ := e.exchange.OrderBook(ctx, e.market.ID, model.OutcomeNo)
	if yes.BestBid != 30 || len(yes.Asks) != 0 {
		t.Fatalf("yes book %+v", yes)
	}
	if no.BestAsk != 70 || len(no.Bids) != 0 {
		t.Fatalf("a YES bid at 30 is a NO ask at 70, got %+v", no)
	}
}

func TestCandles(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	trades := []model.Trade{
		{YesPrice: 50, Size: 1, CreatedAt: base},
		{YesPrice: 60, Size: 2, CreatedAt: base.Add(10 * time.Second)},
		{YesPrice: 40, Size: 3, CreatedAt: base.Add(20 * time.Second)},
		{YesPrice: 45, Size: 1, CreatedAt: base.Add(time.Minute)},
	}
	got := buildCandles(trades, model.OutcomeYes, 100, time.Minute)
	if len(got) != 2 || got[0].Open != 50 || got[0].High != 60 || got[0].Low != 40 || got[0].Close != 40 || got[0].Volume != 6 {
		t.Fatalf("candles %+v", got)
	}
	no := buildCandles(trades, model.OutcomeNo, 100, time.Minute)
	if no[0].Open != 50 || no[0].High != 60 || no[0].Low != 40 || no[0].Close != 60 {
		t.Fatalf("NO candles must mirror: %+v", no)
	}
}

func TestRestingOrderAndFillsAreStreamed(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	if len(e.pub.msgs) != 1 || e.pub.msgs[0].Type != "book" {
		t.Fatalf("a resting order must push the new book, got %+v", e.pub.msgs)
	}
	e.pub.msgs = nil
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))
	if len(e.pub.msgs) != 2 || e.pub.msgs[0].Type != "trade" || e.pub.msgs[1].Type != "book" {
		t.Fatalf("a fill must push the trade then the book, got %+v", e.pub.msgs)
	}
}
