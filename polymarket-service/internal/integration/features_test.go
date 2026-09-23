package integration

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/application"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func TestFeesReferralsAndRewardsOnMySQL(t *testing.T) {
	s := setupWith(t, application.Options{TakerFeeBps: 200, MakerRebateBps: 2000, ReferralBps: 1000, RewardEpoch: time.Minute})
	s.referrals.owner = "carol"
	m := s.newMarket(t, "Fees and rewards")

	var deposited int64
	for _, u := range []string{"alice", "bob", "carol", "operator"} {
		if _, err := s.accounts.Deposit(ctx, u, 100_000); err != nil {
			t.Fatal(err)
		}
		deposited += 100_000
	}

	if _, err := s.referral.Redeem(ctx, "bob", "CODE-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.referral.Redeem(ctx, "bob", "CODE-1"); !errors.Is(err, port.ErrReferralNotAllowed) {
		t.Fatalf("a user is referred once: %v", err)
	}

	if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Buy, 60, 100); err != nil {
		t.Fatal(err)
	}
	res, err := s.place(m.ID, "bob", model.OutcomeNo, model.Buy, 40, 100)
	if err != nil {
		t.Fatal(err)
	}
	tr := res.Trades[0]
	if tr.TakerFee != 80 || tr.MakerRebate != 16 || tr.ReferralFee != 8 {
		t.Fatalf("fee split %+v", tr)
	}
	s.assertSolvent(t, deposited)

	summary, err := s.referral.Summary(ctx, "carol")
	if err != nil || summary.RefCode != "CODE-1" || summary.Referees != 1 || summary.Earnings != 8 {
		t.Fatalf("summary %+v %v", summary, err)
	}
	if bob, _ := s.referral.Summary(ctx, "bob"); bob.ReferredBy != "carol" {
		t.Fatalf("bob %+v", bob)
	}
	if _, err := s.referral.Redeem(ctx, "alice", "CODE-1"); !errors.Is(err, port.ErrReferralNotAllowed) {
		t.Fatalf("alice already traded: %v", err)
	}

	// liquidity rewards, funded by the operator into exchange revenue
	if _, err := s.accounts.FundExchange(ctx, "operator", 10_000); err != nil {
		t.Fatal(err)
	}
	deposited += 10_000
	if _, err := s.rewards.SetMarketRewards(ctx, m.ID, 144_000, 10, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.place(m.ID, "carol", model.OutcomeYes, model.Buy, 48, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Sell, 52, 50); err != nil {
		t.Fatal(err)
	}
	s.assertSolvent(t, deposited)

	epoch := time.Now().Truncate(time.Minute)
	var wg sync.WaitGroup
	var paidTotal atomic.Int64
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() { // six "replicas" race for the same epoch
			defer wg.Done()
			n, err := s.rewards.RunEpoch(ctx, epoch.Add(5*time.Second))
			if err != nil {
				t.Errorf("run epoch: %v", err)
			}
			paidTotal.Add(int64(n))
		}()
	}
	wg.Wait()
	if paidTotal.Load() != 2 {
		t.Fatalf("exactly one caller must pay the epoch, %d payouts made", paidTotal.Load())
	}
	var payouts int64
	s.db.Raw("SELECT COUNT(*) FROM reward_payouts").Scan(&payouts)
	if payouts != 2 {
		t.Fatalf("%d payout rows", payouts)
	}
	s.assertSolvent(t, deposited)

	for i := 1; i <= 3; i++ { // more epochs, then page through them
		if _, err := s.rewards.RunEpoch(ctx, epoch.Add(time.Duration(i)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	ids := walk(t, 2, func(c string) (*port.Page[model.RewardPayout], error) {
		p, _, err := s.rewards.UserRewards(ctx, "carol", 2, c)
		return p, err
	}, func(p model.RewardPayout) int64 { return p.ID })
	assertNoDupes(t, ids, 4)
	if _, total, err := s.rewards.UserRewards(ctx, "carol", 10, ""); err != nil || total.TotalPaid <= 0 {
		t.Fatalf("total %+v %v", total, err)
	}
	s.assertSolvent(t, deposited)

	ex, err := s.accounts.ExchangeSummary(ctx)
	if err != nil || ex.Revenue <= 0 || ex.Revenue >= 10_000+64 {
		t.Fatalf("revenue must be fees + funding minus rewards paid: %+v %v", ex, err)
	}
}

func TestNegRiskConversionAndResolutionOnMySQL(t *testing.T) {
	s := setup(t)
	ev, err := s.catalog.CreateEvent(ctx, port.CreateEventInput{
		Title: "Who wins on mysql", NegRisk: true, EndDate: time.Now().Add(time.Hour),
		Markets: []port.CreateMarketInput{{Question: "A?", GroupItemTitle: "A"}, {Question: "B?", GroupItemTitle: "B"}, {Question: "C?", GroupItemTitle: "C"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	a, b, c := ev.Markets[0].ID, ev.Markets[1].ID, ev.Markets[2].ID

	var deposited int64
	for _, u := range []string{"alice", "bob"} {
		if _, err := s.accounts.Deposit(ctx, u, 100_000); err != nil {
			t.Fatal(err)
		}
		deposited += 100_000
	}
	for _, id := range []int64{a, b} {
		if _, err := s.place(id, "bob", model.OutcomeYes, model.Buy, 40, 10); err != nil {
			t.Fatal(err)
		}
		if _, err := s.place(id, "alice", model.OutcomeNo, model.Buy, 60, 10); err != nil {
			t.Fatal(err)
		}
	}

	res, err := s.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a, b}, Amount: 10})
	if err != nil || res.Cash != 1_000 || len(res.YesMarkets) != 1 || res.YesMarkets[0] != c {
		t.Fatalf("%+v %v", res, err)
	}
	var yesInC int64
	s.db.Raw("SELECT shares FROM positions WHERE market_id = ? AND user_id = 'alice' AND outcome = 'yes'", c).Scan(&yesInC)
	if yesInC != 10 {
		t.Fatalf("alice should hold 10 YES in C, has %d", yesInC)
	}
	if _, err := s.exchange.Convert(ctx, port.ConvertInput{EventID: ev.ID, UserID: "alice", MarketIDs: []int64{a}, Amount: 1}); !errors.Is(err, port.ErrInsufficientShares) {
		t.Fatalf("the NO shares are gone: %v", err)
	}

	if _, err := s.resolution.Resolve(ctx, a, model.OutcomeYes); !errors.Is(err, port.ErrNegRisk) {
		t.Fatalf("%v", err)
	}
	if _, err := s.resolution.ResolveEvent(ctx, ev.ID, c); err != nil {
		t.Fatal(err)
	}
	var cash int64
	s.db.Raw("SELECT SUM(available + locked) FROM balances").Scan(&cash)
	if cash != deposited {
		t.Fatalf("after resolving, all deposits are cash again: %d != %d", cash, deposited)
	}
	if got, _ := s.catalog.GetEvent(ctx, "who-wins-on-mysql"); got.Status != model.EventResolved {
		t.Fatalf("%+v", got)
	}
}

func TestDisputeBondOnMySQL(t *testing.T) {
	s := setupWith(t, application.Options{DisputeBond: 500})
	m := s.newMarket(t, "Bond test")
	var deposited int64
	for _, u := range []string{"alice", "bob", "carol"} {
		if _, err := s.accounts.Deposit(ctx, u, 10_000); err != nil {
			t.Fatal(err)
		}
		deposited += 10_000
	}
	s.place(m.ID, "alice", model.OutcomeYes, model.Buy, 60, 10)
	s.place(m.ID, "bob", model.OutcomeNo, model.Buy, 40, 10)

	if _, err := s.resolution.Propose(ctx, m.ID, model.OutcomeYes); err != nil {
		t.Fatal(err)
	}
	if _, err := s.resolution.Dispute(ctx, m.ID, "nobody", "x"); !errors.Is(err, port.ErrInsufficientBalance) {
		t.Fatalf("%v", err)
	}
	if _, err := s.resolution.Dispute(ctx, m.ID, "carol", "spite"); err != nil {
		t.Fatal(err)
	}
	s.assertSolvent(t, deposited)

	if _, err := s.resolution.Resolve(ctx, m.ID, model.OutcomeYes); err != nil { // proposal stands: bond forfeited
		t.Fatal(err)
	}
	ex, _ := s.accounts.ExchangeSummary(ctx)
	if ex.Revenue != 500 || ex.BondHeld != 0 {
		t.Fatalf("%+v", ex)
	}
	if bal, _ := s.accounts.Balance(ctx, "carol"); bal.Available != 9_500 {
		t.Fatalf("%+v", bal)
	}
	if lb, err := s.social.Leaderboard(ctx, "profit", "day", 10); err != nil || lb[len(lb)-1].UserID != "carol" || lb[len(lb)-1].Profit != -500 {
		t.Fatalf("the forfeited bond is carol's loss: %+v %v", lb, err)
	}
}

func TestProfitLeaderboardHonoursTheTimeWindow(t *testing.T) {
	s := setup(t)
	m := s.newMarket(t, "Windowed profit")
	for _, u := range []string{"alice", "bob"} {
		if _, err := s.accounts.Deposit(ctx, u, 10_000); err != nil {
			t.Fatal(err)
		}
	}
	s.place(m.ID, "alice", model.OutcomeYes, model.Buy, 60, 10)
	s.place(m.ID, "bob", model.OutcomeNo, model.Buy, 40, 10)
	if _, err := s.resolution.Resolve(ctx, m.ID, model.OutcomeYes); err != nil { // alice +400, bob -400
		t.Fatal(err)
	}

	all, err := s.social.Leaderboard(ctx, "profit", "all", 10)
	if err != nil || len(all) != 2 || all[0].UserID != "alice" || all[0].Profit != 400 || all[1].Profit != -400 {
		t.Fatalf("%+v %v", all, err)
	}

	// Push alice's entries eight days back: they leave the day and week windows
	// but stay in the month.
	s.db.Exec("UPDATE pnl_entries SET created_at = ? WHERE user_id = 'alice'", time.Now().Add(-8*24*time.Hour))
	for window, want := range map[string][]string{"day": {"bob"}, "week": {"bob"}, "month": {"alice", "bob"}, "all": {"alice", "bob"}} {
		rows, err := s.social.Leaderboard(ctx, "profit", window, 10)
		if err != nil || len(rows) != len(want) {
			t.Fatalf("%s: %+v %v", window, rows, err)
		}
		for i, u := range want {
			if rows[i].UserID != u {
				t.Fatalf("%s: %+v", window, rows)
			}
		}
	}
	if _, err := s.social.Leaderboard(ctx, "profit", "decade", 10); !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	_ = fmt.Sprint()
}
