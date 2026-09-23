package application

import (
	"math"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
)

func rewardMarket() *model.Market {
	return &model.Market{ID: 1, ShareValue: 100, BestBid: 48, BestAsk: 52, RewardPool: 144_000, RewardMaxSpread: 5, RewardMinSize: 10}
}

func rewardOrder(user string, side model.BookSide, yesPrice, size int64) model.Order {
	return model.Order{UserID: user, BookSide: side, YesPrice: yesPrice, Size: size, Status: model.OrderOpen}
}

func TestRewardScoreFavoursOrdersNearTheMidpoint(t *testing.T) {
	m := rewardMarket() // mid 50, max spread 5
	scores := rewardScores(m, []model.Order{
		rewardOrder("near", model.Bid, 50, 100), // dist 0: weight 1
		rewardOrder("far", model.Bid, 47, 100),  // dist 3: weight (2/5)^2 = 0.16
		rewardOrder("out", model.Bid, 44, 100),  // dist 6: outside
		rewardOrder("small", model.Bid, 50, 5),  // under the minimum size
	}, time.Now())

	if math.Abs(scores["near"]-100/3.0) > 1e-9 {
		t.Fatalf("one-sided score is a third of the raw score, got %v", scores["near"])
	}
	if math.Abs(scores["far"]-16/3.0) > 1e-9 {
		t.Fatalf("far %v", scores["far"])
	}
	if _, ok := scores["out"]; ok {
		t.Fatal("orders beyond max spread score nothing")
	}
	if _, ok := scores["small"]; ok {
		t.Fatal("orders under the minimum size score nothing")
	}
}

func TestTwoSidedQuotingScoresBothSides(t *testing.T) {
	scores := rewardScores(rewardMarket(), []model.Order{
		rewardOrder("mm", model.Bid, 49, 100), rewardOrder("mm", model.Ask, 51, 100),
		rewardOrder("one", model.Bid, 49, 100),
	}, time.Now())
	w := (5.0 - 1.0) / 5.0
	if math.Abs(scores["mm"]-2*w*w*100) > 1e-9 || math.Abs(scores["one"]-w*w*100/3) > 1e-9 {
		t.Fatalf("%v", scores)
	}
}

func TestOneSidedQuotingScoresNothingNearTheExtremes(t *testing.T) {
	m := rewardMarket()
	m.BestBid, m.BestAsk = 4, 8 // mid 6% < 10%
	scores := rewardScores(m, []model.Order{
		rewardOrder("one", model.Bid, 6, 100),
		rewardOrder("two", model.Bid, 5, 100), rewardOrder("two", model.Ask, 7, 100),
	}, time.Now())
	if _, ok := scores["one"]; ok {
		t.Fatal("one-sided quoting near the extremes earns nothing")
	}
	if scores["two"] <= 0 {
		t.Fatal("two-sided quoting still earns")
	}
}

func TestExpiredOrdersDoNotScore(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	o := rewardOrder("u", model.Bid, 50, 100)
	o.ExpiresAt = &past
	if s := rewardScores(rewardMarket(), []model.Order{o}, time.Now()); len(s) != 0 {
		t.Fatalf("%v", s)
	}
}

// rewardedEnv has a market paying 144_000/day (100 per one-minute epoch) and
// two makers quoting either side, so revenue is the only limit.
func rewardedEnv(t *testing.T) *env {
	t.Helper()
	e := newEnvWith(Options{RewardEpoch: time.Minute})
	if _, err := e.rewards.SetMarketRewards(ctx, e.market.ID, 144_000, 5, 10); err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{"alice", "bob", "carol"} {
		e.fund(u, 100_000)
	}
	// alice and bob mint a position to trade against, carol quotes both sides tightly
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 100))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 50, 100))
	e.must(e.limit("carol", model.OutcomeYes, model.Buy, 48, 100))
	e.must(e.limit("alice", model.OutcomeYes, model.Sell, 52, 100))
	return e
}

func TestRewardEpochPaysMakersFromExchangeRevenue(t *testing.T) {
	e := rewardedEnv(t)
	now := time.Now()

	if n, err := e.rewards.RunEpoch(ctx, now); err != nil || n != 0 {
		t.Fatalf("with no revenue nothing can be paid: n=%d err=%v", n, err)
	}

	e.wallets.wallet.ID = "wallet-operator"
	if _, err := e.accounts.FundExchange(ctx, "operator", 10_000); err != nil {
		t.Fatal(err)
	}
	e.deposited += 10_000

	n, err := e.rewards.RunEpoch(ctx, now.Add(time.Minute))
	if err != nil || n != 2 {
		t.Fatalf("carol and alice quote near the midpoint: n=%d err=%v", n, err)
	}
	var paid int64
	for _, p := range e.mem.payouts {
		paid += p.Amount
		if p.Amount <= 0 {
			t.Fatalf("payout %+v", p)
		}
	}
	if paid <= 0 || paid > 100 {
		t.Fatalf("an epoch pays at most pool/1440 = 100, paid %d", paid)
	}
	if e.revenue() != 10_000-paid {
		t.Fatalf("payouts come out of revenue: revenue %d paid %d", e.revenue(), paid)
	}
	e.checkConservation(t)
}

func TestRewardEpochIsPaidOnce(t *testing.T) {
	e := rewardedEnv(t)
	e.mem.exch = append(e.mem.exch, model.ExchangeEntry{Bucket: model.BucketRevenue, Kind: "funding", Amount: 10_000})
	e.deposited += 10_000
	now := time.Now().Truncate(time.Minute).Add(10 * time.Second)

	first, _ := e.rewards.RunEpoch(ctx, now)
	second, _ := e.rewards.RunEpoch(ctx, now.Add(20*time.Second)) // same epoch, e.g. another replica
	if first == 0 || second != 0 {
		t.Fatalf("first %d second %d: exactly one caller pays an epoch", first, second)
	}
}

func TestRewardsNeverExceedRevenue(t *testing.T) {
	e := rewardedEnv(t)
	e.mem.exch = append(e.mem.exch, model.ExchangeEntry{Bucket: model.BucketRevenue, Kind: "funding", Amount: 7})
	e.deposited += 7
	if _, err := e.rewards.RunEpoch(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	if e.revenue() < 0 {
		t.Fatalf("revenue went negative: %d", e.revenue())
	}
	var paid int64
	for _, p := range e.mem.payouts {
		paid += p.Amount
	}
	if paid > 7 {
		t.Fatalf("paid %d out of 7", paid)
	}
}

func TestRewardConfigurationIsValidated(t *testing.T) {
	e := newEnv()
	for name, args := range map[string][3]int64{
		"pool without spread": {100, 0, 1},
		"spread too wide":     {100, 100, 1},
		"negative pool":       {-1, 5, 1},
	} {
		if _, err := e.rewards.SetMarketRewards(ctx, e.market.ID, args[0], args[1], args[2]); err == nil {
			t.Errorf("%s should be rejected", name)
		}
	}
	if _, err := e.rewards.SetMarketRewards(ctx, e.market.ID, 0, 0, 0); err != nil {
		t.Fatalf("turning rewards off: %v", err)
	}
}
