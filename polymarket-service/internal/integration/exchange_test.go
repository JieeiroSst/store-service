// Package integration runs the real repositories and use cases against MySQL.
// It is skipped unless MYSQL_TEST_DSN is set, e.g.
//
//	MYSQL_TEST_DSN='root:pw@tcp(127.0.0.1:3306)/pm?charset=utf8mb4&parseTime=True&loc=Local' go test ./internal/integration
package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/realtime"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/polymarket-service/internal/application"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/JIeeiroSst/polymarket-service/internal/infrastructure/database"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ctx = context.Background()

type wallets struct{ n atomic.Int64 }

func (w *wallets) GetWalletByUser(context.Context, string) (*port.Wallet, error) {
	return &port.Wallet{ID: "w-user", Currency: "USD", Status: "active"}, nil
}
func (w *wallets) Transfer(context.Context, port.TransferInput) (string, error) {
	return fmt.Sprintf("tr-%d", w.n.Add(1)), nil
}
func (w *wallets) ReverseTransfer(context.Context, string, string) error { return nil }

type notifier struct{}

func (notifier) Notify(context.Context, port.Notification) error { return nil }

type stack struct {
	db         *gorm.DB
	catalog    port.CatalogUsecase
	exchange   port.ExchangeUsecase
	resolution port.ResolutionUsecase
	accounts   port.AccountUsecase
	social     port.SocialUsecase
	rewards    port.RewardUsecase
	referral   port.ReferralUsecase
	referrals  *referralGateway
}

type referralGateway struct{ owner string }

func (g *referralGateway) GenerateLink(context.Context, string) (*port.ReferralLink, error) {
	return &port.ReferralLink{RefCode: "CODE-1", DeepLink: "app://ref"}, nil
}
func (g *referralGateway) Activate(context.Context, string, string) (string, error) {
	return g.owner, nil
}
func (g *referralGateway) Stats(context.Context, string) (*port.ReferralStats, error) {
	return &port.ReferralStats{}, nil
}

func setup(t *testing.T) *stack { return setupWith(t, application.Options{}) }

func setupWith(t *testing.T, opts application.Options) *stack {
	t.Helper()
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range []string{"events", "markets", "orders", "trades", "balances", "ledger_entries", "positions", "profiles", "comments", "comment_likes", "bookmarks", "pnl_entries", "exchange_entries", "reward_epochs", "reward_payouts", "referral_codes", "referrals"} {
		db.Exec("DROP TABLE IF EXISTS " + tbl)
	}
	if err := database.ApplySchema(db, "../../database.sql"); err != nil {
		t.Fatal(err)
	}
	// Applying twice must be harmless: it happens on every startup.
	if err := database.ApplySchema(db, "../../database.sql"); err != nil {
		t.Fatalf("schema is not idempotent: %v", err)
	}

	tx := repository.NewTxManager(db)
	events, markets := repository.NewEventRepository(db), repository.NewMarketRepository(db)
	orders, trades := repository.NewOrderRepository(db), repository.NewTradeRepository(db)
	balances, ledger := repository.NewBalanceRepository(db), repository.NewLedgerRepository(db)
	positions, socialRepo := repository.NewPositionRepository(db), repository.NewSocialRepository(db)
	stats, pnl := repository.NewStatsRepository(db), repository.NewPnLRepository(db)
	exchange, rewardRepo := repository.NewExchangeAccountRepository(db), repository.NewRewardRepository(db)
	referrals := repository.NewReferralRepository(db)
	hub, n, w := realtime.NewHub(), notifier{}, &wallets{}
	gateway := &referralGateway{owner: "ref-owner"}
	opts.TreasuryWalletID, opts.Currency, opts.ShareValue = "treasury", "USD", 100
	if opts.DisputeWindow == 0 {
		opts.DisputeWindow = time.Hour
	}

	return &stack{
		db:      db,
		catalog: application.NewCatalogService(events, markets, tx, opts),
		exchange: application.NewExchangeService(application.ExchangeParams{
			Markets: markets, Events: events, Orders: orders, Trades: trades, Balances: balances, Positions: positions,
			Ledger: ledger, Referrals: referrals, PnL: pnl, Exchange: exchange, Tx: tx, Notifier: n, Publisher: hub, Opts: opts,
		}),
		resolution: application.NewResolutionService(application.ResolutionParams{
			Markets: markets, Events: events, Orders: orders, Positions: positions, Balances: balances, Ledger: ledger,
			Referrals: referrals, PnL: pnl, Exchange: exchange, Tx: tx, Notifier: n, Publisher: hub, Opts: opts,
		}),
		accounts: application.NewAccountService(application.AccountParams{
			Markets: markets, Trades: trades, Balances: balances, Ledger: ledger, Positions: positions, Exchange: exchange,
			Tx: tx, Wallets: w, Notifier: n, Opts: opts,
		}),
		social: application.NewSocialService(socialRepo, events, markets, positions, stats),
		rewards: application.NewRewardService(application.RewardParams{
			Markets: markets, Orders: orders, Positions: positions, Balances: balances, Referrals: referrals, PnL: pnl,
			Exchange: exchange, Rewards: rewardRepo, Tx: tx, Notifier: n, Opts: opts,
		}),
		referral: application.NewReferralService(application.ReferralParams{
			Referrals: referrals, Gateway: gateway, Trades: trades, Notifier: n, Opts: opts,
		}),
		referrals: gateway,
	}
}

func (s *stack) newMarket(t *testing.T, title string) model.Market {
	t.Helper()
	ev, err := s.catalog.CreateEvent(ctx, port.CreateEventInput{
		Title: title, Category: "Politics", Tags: []string{"US", "election"}, EndDate: time.Now().Add(time.Hour),
		Markets: []port.CreateMarketInput{{Question: title + "?"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return ev.Markets[0]
}

func (s *stack) place(marketID int64, user string, o model.Outcome, side model.Side, price, size int64) (*port.PlaceOrderResult, error) {
	return s.exchange.PlaceOrder(ctx, port.PlaceOrderInput{MarketID: marketID, UserID: user, Outcome: o, Side: side, Price: price, Size: size})
}

// assertSolvent checks the ledger identity directly in SQL: cash in balances
// plus one share value per outstanding YES/NO pair equals what was deposited.
func (s *stack) assertSolvent(t *testing.T, deposited int64) {
	t.Helper()
	var cash, yes, no, locked, orderLocked, reserve, exchange int64
	s.db.Raw("SELECT COALESCE(SUM(available + locked),0) FROM balances").Scan(&cash)
	s.db.Raw("SELECT COALESCE(SUM(locked),0) FROM balances").Scan(&locked)
	s.db.Raw("SELECT COALESCE(SUM(locked_cash),0) FROM orders WHERE status='open'").Scan(&orderLocked)
	s.db.Raw("SELECT COALESCE(SUM(shares),0) FROM positions WHERE outcome='yes'").Scan(&yes)
	s.db.Raw("SELECT COALESCE(SUM(shares),0) FROM positions WHERE outcome='no'").Scan(&no)
	s.db.Raw("SELECT COALESCE(SUM(fee_reserve),0) FROM orders WHERE status='open'").Scan(&reserve)
	orderLocked += reserve
	if locked != orderLocked {
		t.Fatalf("balances.locked %d != open orders' locked_cash %d", locked, orderLocked)
	}
	if yes != no {
		t.Fatalf("YES supply %d != NO supply %d", yes, no)
	}
	s.db.Raw("SELECT COALESCE(SUM(amount),0) FROM exchange_entries").Scan(&exchange)
	if cash+exchange+yes*100 != deposited {
		t.Fatalf("cash %d + exchange %d + backing %d != deposited %d", cash, exchange, yes*100, deposited)
	}
}

func TestFullLifecycleOnMySQL(t *testing.T) {
	s := setup(t)
	m := s.newMarket(t, "Will X win")

	var deposited int64
	for _, u := range []string{"alice", "bob", "carol"} {
		if _, err := s.accounts.Deposit(ctx, u, 10_000); err != nil {
			t.Fatal(err)
		}
		deposited += 10_000
	}

	// mint
	if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Buy, 60, 10); err != nil {
		t.Fatal(err)
	}
	res, err := s.place(m.ID, "bob", model.OutcomeNo, model.Buy, 40, 10)
	if err != nil || len(res.Trades) != 1 || res.Trades[0].Kind != model.KindMint {
		t.Fatalf("mint: %+v %v", res, err)
	}
	// swap: alice sells to carol
	if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Sell, 65, 10); err != nil {
		t.Fatal(err)
	}
	res, err = s.place(m.ID, "carol", model.OutcomeYes, model.Buy, 70, 4)
	if err != nil || len(res.Trades) != 1 || res.Trades[0].YesPrice != 65 || res.Trades[0].Kind != model.KindSwap {
		t.Fatalf("swap: %+v %v", res, err)
	}
	s.assertSolvent(t, deposited)

	// book, top of book and views
	book, err := s.exchange.OrderBook(ctx, m.ID, model.OutcomeYes)
	if err != nil || len(book.Asks) != 1 || book.Asks[0].Size != 6 || book.BestAsk != 65 {
		t.Fatalf("book %+v %v", book, err)
	}
	got, _ := s.catalog.GetMarket(ctx, fmt.Sprint(m.ID))
	if got.BestAsk != 65 || got.LastPrice != 65 || got.Volume != 60*10+65*4 {
		t.Fatalf("market top of book %+v", got)
	}
	if candles, err := s.exchange.PriceHistory(ctx, port.PriceHistoryInput{MarketID: m.ID, Outcome: model.OutcomeYes, Range: "1h"}); err != nil || len(candles) == 0 {
		t.Fatalf("candles %+v %v", candles, err)
	}

	// portfolio, activity, holders, leaderboard
	pf, err := s.accounts.Portfolio(ctx, "carol")
	if err != nil || len(pf.Positions) != 1 || pf.Positions[0].Shares != 4 {
		t.Fatalf("portfolio %+v %v", pf, err)
	}
	if act, err := s.accounts.Activity(ctx, "alice", 10); err != nil || len(act) < 3 {
		t.Fatalf("activity %+v %v", act, err)
	}
	if h, err := s.social.TopHolders(ctx, m.ID, model.OutcomeYes, 5); err != nil || len(h) == 0 {
		t.Fatalf("holders %+v %v", h, err)
	}
	if lb, err := s.social.Leaderboard(ctx, "volume", "all", 10); err != nil || len(lb) == 0 || lb[0].Volume == 0 {
		t.Fatalf("volume leaderboard %+v %v", lb, err)
	}

	// cancel then resolve
	open, _ := s.exchange.ListOrders(ctx, port.OrderFilter{UserID: "alice", Status: model.OrderOpen})
	if len(open.Items) != 1 || !open.IsLastPage {
		t.Fatalf("alice should have 1 open order, has %+v", open)
	}
	if _, err := s.exchange.CancelOrder(ctx, open.Items[0].ID, "alice"); err != nil {
		t.Fatal(err)
	}
	s.assertSolvent(t, deposited)

	if _, err := s.place(m.ID, "carol", model.OutcomeNo, model.Buy, 30, 5); err != nil { // resting, refunded on resolve
		t.Fatal(err)
	}
	if _, err := s.resolution.Resolve(ctx, m.ID, model.OutcomeYes); err != nil {
		t.Fatal(err)
	}
	var cash int64
	s.db.Raw("SELECT SUM(available + locked) FROM balances").Scan(&cash)
	if cash != deposited {
		t.Fatalf("after resolution all deposits must be cash again: %d != %d", cash, deposited)
	}
	if lb, err := s.social.Leaderboard(ctx, "profit", "", 10); err != nil || len(lb) == 0 {
		t.Fatalf("profit leaderboard %+v %v", lb, err)
	}
	ev, _ := s.catalog.GetEvent(ctx, "will-x-win")
	if ev.Status != model.EventResolved || ev.Markets[0].Status != model.MarketResolved {
		t.Fatalf("event %+v", ev)
	}
	if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Buy, 50, 1); !errors.Is(err, port.ErrMarketNotTradable) {
		t.Fatalf("resolved markets are closed: %v", err)
	}
}

func TestClientOrderIDsAreOnlyUniquePerUserWhenSet(t *testing.T) {
	s := setup(t)
	m := s.newMarket(t, "Idempotency")
	if _, err := s.accounts.Deposit(ctx, "alice", 10_000); err != nil {
		t.Fatal(err)
	}
	// Several orders without a client id must not collide.
	for i := 0; i < 3; i++ {
		if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Buy, int64(10+i), 1); err != nil {
			t.Fatalf("order %d: %v", i, err)
		}
	}
	in := port.PlaceOrderInput{MarketID: m.ID, UserID: "alice", ClientOrderID: "c1", Outcome: model.OutcomeYes, Side: model.Buy, Price: 20, Size: 1}
	a, err := s.exchange.PlaceOrder(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.exchange.PlaceOrder(ctx, in)
	if err != nil || a.Order.ID != b.Order.ID {
		t.Fatalf("retry must return the same order: %+v %v", b, err)
	}
}

func TestCatalogAndSocial(t *testing.T) {
	s := setup(t)
	m := s.newMarket(t, "Catalog check")
	if _, err := s.catalog.CreateEvent(ctx, port.CreateEventInput{
		Title: "Catalog check", EndDate: time.Now().Add(time.Hour), Markets: []port.CreateMarketInput{{Question: "again"}},
	}); !errors.Is(err, port.ErrAlreadyExists) {
		t.Fatalf("duplicate slug: %v", err)
	}

	multi, err := s.catalog.CreateEvent(ctx, port.CreateEventInput{
		Title: "Who wins", Category: "politics", Featured: true, NegRisk: true, EndDate: time.Now().Add(time.Hour),
		Markets: []port.CreateMarketInput{{Question: "Alice wins?", GroupItemTitle: "Alice"}, {Question: "Bob wins?", GroupItemTitle: "Bob"}},
	})
	if err != nil || len(multi.Markets) != 2 {
		t.Fatalf("multi-outcome event: %+v %v", multi, err)
	}

	list, err := s.catalog.ListEvents(ctx, port.EventFilter{Tag: "election", Category: "politics", Limit: 10})
	if err != nil || len(list.Items) != 1 || !list.IsLastPage || len(list.Items[0].Markets) == 0 {
		t.Fatalf("tag+category filter: %+v %v", list, err)
	}
	for _, f := range []port.EventFilter{{Query: "wins"}, {Featured: true}, {Sort: "newest"}, {Sort: "ending"}, {Status: model.EventOpen}} {
		f.Limit = 10
		if r, err := s.catalog.ListEvents(ctx, f); err != nil || len(r.Items) == 0 {
			t.Fatalf("filter %+v: %+v %v", f, r, err)
		}
	}
	if ms, err := s.catalog.SearchMarkets(ctx, port.MarketFilter{Query: "Bob", Limit: 5}); err != nil || ms.Total != 1 {
		t.Fatalf("market search %+v %v", ms, err)
	}
	if cats, err := s.catalog.Categories(ctx); err != nil || len(cats) != 1 {
		t.Fatalf("categories %+v %v", cats, err)
	}
	if rel, err := s.catalog.RelatedEvents(ctx, "who-wins", 5); err != nil {
		t.Fatalf("related %+v %v", rel, err)
	}

	// comments, likes, replies
	c1, err := s.social.AddComment(ctx, multi.ID, "alice", "first!", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.social.AddComment(ctx, multi.ID, "bob", "reply", c1.ID); err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{"bob", "bob", "carol"} { // duplicate like is a no-op
		if err := s.social.LikeComment(ctx, c1.ID, u); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.social.ListComments(ctx, multi.ID, "likes", 10, "")
	if err != nil || len(page.Items) != 2 || page.Items[0].Likes != 2 || !page.IsLastPage {
		t.Fatalf("comments %+v %v", page, err)
	}
	if err := s.social.DeleteComment(ctx, c1.ID, "bob"); !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("only the author may delete: %v", err)
	}
	if err := s.social.DeleteComment(ctx, c1.ID, "alice"); err != nil {
		t.Fatal(err)
	}
	if left, _ := s.social.ListComments(ctx, multi.ID, "", 10, ""); len(left.Items) != 0 {
		t.Fatalf("deleting a comment must delete its replies, %d left", len(left.Items))
	}

	// watchlist
	if err := s.social.Watch(ctx, "alice", multi.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.social.Watch(ctx, "alice", multi.ID); err != nil {
		t.Fatalf("watching twice is a no-op: %v", err)
	}
	if wl, err := s.social.Watchlist(ctx, "alice"); err != nil || len(wl) != 1 {
		t.Fatalf("watchlist %+v %v", wl, err)
	}

	// profiles
	if _, err := s.social.UpdateProfile(ctx, port.UpdateProfileInput{UserID: "alice", Username: "alice_w", Bio: "hi"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.social.UpdateProfile(ctx, port.UpdateProfileInput{UserID: "bob", Username: "alice_w"}); !errors.Is(err, port.ErrAlreadyExists) {
		t.Fatalf("usernames are unique: %v", err)
	}
	if p, err := s.social.GetProfile(ctx, "alice"); err != nil || p.Username != "alice_w" || p.Bio != "hi" {
		t.Fatalf("a rejected username must not touch the owner's profile: %+v %v", p, err)
	}
	if _, err := s.social.UpdateProfile(ctx, port.UpdateProfileInput{UserID: "alice", Username: "alice_w", Bio: "hi"}); err != nil {
		t.Fatalf("saving unchanged values must work: %v", err)
	}
	if _, err := s.social.UpdateProfile(ctx, port.UpdateProfileInput{UserID: "carol"}); err == nil {
		t.Fatal("a profile needs a username")
	}
	_ = m
}

// TestConcurrentTradingKeepsLedgerConsistent hammers one market from many
// users at once: the ledger identities must hold and no order may leak escrow.
func TestConcurrentTradingKeepsLedgerConsistent(t *testing.T) {
	// Fees, rebates and referral commissions spread each fill over more balance
	// rows (maker, taker, referrer), which is where deadlocks would come from.
	s := setupWith(t, application.Options{TakerFeeBps: 200, MakerRebateBps: 2000, ReferralBps: 1000})
	m := s.newMarket(t, "Concurrency")
	for i := 0; i < 8; i += 2 {
		s.db.Exec("INSERT INTO referrals (referee_user_id, referrer_user_id, ref_code) VALUES (?, ?, 'X')", fmt.Sprintf("u%d", i), fmt.Sprintf("u%d", (i+3)%8))
	}

	const users = 8
	var deposited int64
	for i := 0; i < users; i++ {
		if _, err := s.accounts.Deposit(ctx, fmt.Sprintf("u%d", i), 100_000); err != nil {
			t.Fatal(err)
		}
		deposited += 100_000
	}

	var wg sync.WaitGroup
	var failed atomic.Int64
	for i := 0; i < users; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			user := fmt.Sprintf("u%d", i)
			for j := 0; j < 25; j++ {
				outcome, side := model.OutcomeYes, model.Buy
				if (i+j)%2 == 1 {
					outcome = model.OutcomeNo
				}
				price := int64(40 + (i*7+j*3)%20)
				if _, err := s.place(m.ID, user, outcome, side, price, int64(1+j%4)); err != nil {
					failed.Add(1)
					t.Errorf("place: %v", err)
					return
				}
				if j%5 == 4 { // and sometimes sell what we hold
					_, _ = s.place(m.ID, user, outcome, model.Sell, price+5, 1)
				}
			}
		}(i)
	}
	wg.Wait()
	if failed.Load() > 0 {
		t.FailNow()
	}
	s.assertSolvent(t, deposited)

	var trades int64
	s.db.Raw("SELECT COUNT(*) FROM trades").Scan(&trades)
	if trades == 0 {
		t.Fatal("expected some trades to have crossed")
	}
	if _, err := s.resolution.Resolve(ctx, m.ID, model.OutcomeNo); err != nil {
		t.Fatal(err)
	}
	var cash, locked, exchange int64
	s.db.Raw("SELECT SUM(available + locked), SUM(locked) FROM balances").Row().Scan(&cash, &locked)
	s.db.Raw("SELECT COALESCE(SUM(amount), 0) FROM exchange_entries").Scan(&exchange)
	if cash+exchange != deposited || locked != 0 {
		t.Fatalf("after resolution cash=%d exchange=%d locked=%d, deposited=%d", cash, exchange, locked, deposited)
	}
	if exchange <= 0 {
		t.Fatal("fees should have produced exchange revenue")
	}
}

// walk follows next_cursor until is_last_page, asserting the flag and cursor
// agree on every page, and returns the ids in the order they arrived.
func walk[T any](t *testing.T, size int, fetch func(cursor string) (*port.Page[T], error), id func(T) int64) []int64 {
	t.Helper()
	var ids []int64
	cursor := ""
	for pages := 0; ; pages++ {
		if pages > 50 {
			t.Fatal("pagination does not terminate")
		}
		p, err := fetch(cursor)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Items) > size {
			t.Fatalf("page has %d items, limit %d", len(p.Items), size)
		}
		for _, it := range p.Items {
			ids = append(ids, id(it))
		}
		if p.IsLastPage != (p.NextCursor == "") {
			t.Fatalf("is_last_page=%v but next_cursor=%q", p.IsLastPage, p.NextCursor)
		}
		if p.IsLastPage {
			return ids
		}
		if len(p.Items) != size {
			t.Fatalf("a non-last page must be full: %d of %d", len(p.Items), size)
		}
		cursor = p.NextCursor
	}
}

func assertNoDupes(t *testing.T, ids []int64, want int) {
	t.Helper()
	seen := map[int64]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("id %d returned twice: %v", id, ids)
		}
		seen[id] = true
	}
	if len(ids) != want {
		t.Fatalf("got %d ids, want %d: %v", len(ids), want, ids)
	}
}

func TestCursorPagination(t *testing.T) {
	s := setup(t)

	// 7 events; give some of them volume through real trades so the volume sort has ties and gaps.
	var markets []model.Market
	for i := 0; i < 7; i++ {
		markets = append(markets, s.newMarket(t, fmt.Sprintf("Pagination %d", i)))
	}
	for _, u := range []string{"alice", "bob"} {
		if _, err := s.accounts.Deposit(ctx, u, 100_000); err != nil {
			t.Fatal(err)
		}
	}
	for i, size := range []int64{5, 5, 9, 0, 0, 2, 9} { // ties at 5 and 9, zero-volume ties
		if size == 0 {
			continue
		}
		s.place(markets[i].ID, "alice", model.OutcomeYes, model.Buy, 50, size)
		if _, err := s.place(markets[i].ID, "bob", model.OutcomeNo, model.Buy, 50, size); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("events by volume", func(t *testing.T) {
		var vols []int64
		ids := walk(t, 3, func(c string) (*port.Page[model.Event], error) {
			p, err := s.catalog.ListEvents(ctx, port.EventFilter{Limit: 3, Cursor: c})
			if p != nil {
				for _, e := range p.Items {
					vols = append(vols, e.Volume)
				}
			}
			return p, err
		}, func(e model.Event) int64 { return e.ID })
		assertNoDupes(t, ids, 7)
		for i := 1; i < len(vols); i++ {
			if vols[i] > vols[i-1] {
				t.Fatalf("volumes must be non-increasing: %v", vols)
			}
		}
	})

	t.Run("events newest and ending", func(t *testing.T) {
		for _, sort := range []string{"newest", "ending"} {
			ids := walk(t, 2, func(c string) (*port.Page[model.Event], error) {
				return s.catalog.ListEvents(ctx, port.EventFilter{Limit: 2, Sort: sort, Cursor: c})
			}, func(e model.Event) int64 { return e.ID })
			assertNoDupes(t, ids, 7)
			if sort == "newest" && ids[0] < ids[len(ids)-1] {
				t.Fatalf("newest first expected: %v", ids)
			}
		}
	})

	t.Run("exact multiple of the page size ends cleanly", func(t *testing.T) {
		p, err := s.catalog.ListEvents(ctx, port.EventFilter{Limit: 7})
		if err != nil || len(p.Items) != 7 || !p.IsLastPage || p.NextCursor != "" {
			t.Fatalf("7 items at limit 7 is one last page: %+v %v", p, err)
		}
		p, err = s.catalog.ListEvents(ctx, port.EventFilter{Limit: 6})
		if err != nil || p.IsLastPage || len(p.Items) != 6 {
			t.Fatalf("7 items at limit 6 is not the last page: %+v %v", p, err)
		}
	})

	t.Run("empty result is a last page", func(t *testing.T) {
		p, err := s.catalog.ListEvents(ctx, port.EventFilter{Query: "no such thing", Limit: 5})
		if err != nil || len(p.Items) != 0 || !p.IsLastPage || p.NextCursor != "" {
			t.Fatalf("%+v %v", p, err)
		}
	})

	t.Run("bad and mismatched cursors are rejected", func(t *testing.T) {
		if _, err := s.catalog.ListEvents(ctx, port.EventFilter{Limit: 2, Cursor: "not-a-cursor"}); !errors.Is(err, port.ErrInvalidInput) {
			t.Fatalf("garbage cursor: %v", err)
		}
		first, _ := s.catalog.ListEvents(ctx, port.EventFilter{Limit: 2, Sort: "newest"})
		if _, err := s.catalog.ListEvents(ctx, port.EventFilter{Limit: 2, Sort: "ending", Cursor: first.NextCursor}); !errors.Is(err, port.ErrInvalidInput) {
			t.Fatalf("a cursor must not cross sort orders: %v", err)
		}
	})

	t.Run("user orders", func(t *testing.T) {
		m := markets[3] // no trades yet
		for i := 0; i < 7; i++ {
			if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Buy, int64(10+i), 1); err != nil {
				t.Fatal(err)
			}
		}
		ids := walk(t, 3, func(c string) (*port.Page[model.Order], error) {
			return s.exchange.ListOrders(ctx, port.OrderFilter{UserID: "alice", MarketID: m.ID, Limit: 3, Cursor: c})
		}, func(o model.Order) int64 { return o.ID })
		assertNoDupes(t, ids, 7)
		for i := 1; i < len(ids); i++ {
			if ids[i] > ids[i-1] {
				t.Fatalf("newest first expected: %v", ids)
			}
		}
		// orders created while a client is mid-way must not shift the remaining pages
		first, _ := s.exchange.ListOrders(ctx, port.OrderFilter{UserID: "alice", MarketID: m.ID, Limit: 3})
		if _, err := s.place(m.ID, "alice", model.OutcomeYes, model.Buy, 40, 1); err != nil {
			t.Fatal(err)
		}
		second, err := s.exchange.ListOrders(ctx, port.OrderFilter{UserID: "alice", MarketID: m.ID, Limit: 3, Cursor: first.NextCursor})
		if err != nil || second.Items[0].ID != first.Items[2].ID-1 {
			t.Fatalf("a new order must not shift page 2: %+v %v", second, err)
		}
	})

	t.Run("comments", func(t *testing.T) {
		ev, _ := s.catalog.GetEvent(ctx, "pagination-0")
		var cs []*model.Comment
		for i := 0; i < 7; i++ {
			c, err := s.social.AddComment(ctx, ev.ID, "alice", fmt.Sprintf("c%d", i), 0)
			if err != nil {
				t.Fatal(err)
			}
			cs = append(cs, c)
		}
		// likes: c1 and c4 get 2, c2 gets 1, so the likes order is c4 c1 c2 c6 c5 c3 c0
		for _, u := range []string{"a", "b"} {
			_ = s.social.LikeComment(ctx, cs[1].ID, u)
			_ = s.social.LikeComment(ctx, cs[4].ID, u)
		}
		_ = s.social.LikeComment(ctx, cs[2].ID, "a")

		newest := walk(t, 3, func(c string) (*port.Page[model.Comment], error) {
			return s.social.ListComments(ctx, ev.ID, "newest", 3, c)
		}, func(c model.Comment) int64 { return c.ID })
		assertNoDupes(t, newest, 7)
		if newest[0] != cs[6].ID || newest[6] != cs[0].ID {
			t.Fatalf("newest order wrong: %v", newest)
		}

		liked := walk(t, 3, func(c string) (*port.Page[model.Comment], error) {
			return s.social.ListComments(ctx, ev.ID, "likes", 3, c)
		}, func(c model.Comment) int64 { return c.ID })
		assertNoDupes(t, liked, 7)
		want := []int64{cs[4].ID, cs[1].ID, cs[2].ID, cs[6].ID, cs[5].ID, cs[3].ID, cs[0].ID}
		for i := range want {
			if liked[i] != want[i] {
				t.Fatalf("likes order: got %v want %v", liked, want)
			}
		}
	})
}
