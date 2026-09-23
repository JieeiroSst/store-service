package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

// mem is an in-memory implementation of every driven repository port. Its
// TxManager snapshots state and restores it on error, so rollback behaviour
// (fill-or-kill, failed transfers) is tested for real.
type mem struct {
	nextID    int64
	markets   map[int64]model.Market
	events    map[int64]model.Event
	orders    map[int64]model.Order
	trades    []model.Trade
	balances  map[string]model.Balance
	ledger    []model.LedgerEntry
	pos       map[string]model.Position
	pnl       []model.PnLEntry
	exch      []model.ExchangeEntry
	epochs    map[time.Time]bool
	payouts   []model.RewardPayout
	referrals map[string]model.Referral
	codes     map[string]model.ReferralCode
	failTx    error // returned after fn succeeds, simulating a failed commit
	clock     time.Time
}

func newMem() *mem {
	return &mem{
		markets: map[int64]model.Market{}, events: map[int64]model.Event{}, orders: map[int64]model.Order{},
		balances: map[string]model.Balance{}, pos: map[string]model.Position{}, clock: time.Now(),
		epochs: map[time.Time]bool{}, referrals: map[string]model.Referral{}, codes: map[string]model.ReferralCode{},
	}
}

func (m *mem) id() int64 { m.nextID++; return m.nextID }

func (m *mem) stamp() time.Time { m.clock = m.clock.Add(time.Millisecond); return m.clock }

func posKey(marketID int64, user string, o model.Outcome) string {
	return fmt.Sprintf("%d/%s/%s", marketID, user, o)
}

// ---- TxManager ----

func (m *mem) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	markets, events, orders := copyMap(m.markets), copyMap(m.events), copyMap(m.orders)
	balances, pos := copyMap(m.balances), copyMap(m.pos)
	trades, ledger := append([]model.Trade(nil), m.trades...), append([]model.LedgerEntry(nil), m.ledger...)
	pnl, exch, payouts := append([]model.PnLEntry(nil), m.pnl...), append([]model.ExchangeEntry(nil), m.exch...), append([]model.RewardPayout(nil), m.payouts...)
	referrals := copyMap(m.referrals)

	err := fn(ctx)
	if err == nil {
		err = m.failTx
	}
	if err != nil {
		m.markets, m.events, m.orders, m.balances, m.pos, m.trades, m.ledger = markets, events, orders, balances, pos, trades, ledger
		m.pnl, m.exch, m.payouts, m.referrals = pnl, exch, payouts, referrals
	}
	return err
}

func copyMap[K comparable, V any](in map[K]V) map[K]V {
	out := make(map[K]V, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// ---- MarketRepository ----

func (m *mem) Create(ctx context.Context, x *model.Market) error {
	x.ID = m.id()
	m.markets[x.ID] = *x
	return nil
}

func (m *mem) GetByID(_ context.Context, id int64) (*model.Market, error) {
	x, ok := m.markets[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &x, nil
}

func (m *mem) GetByIDForUpdate(ctx context.Context, id int64) (*model.Market, error) {
	return m.GetByID(ctx, id)
}

func (m *mem) GetBySlug(context.Context, string) (*model.Market, error) { return nil, port.ErrNotFound }

func (m *mem) GetByIDs(_ context.Context, ids []int64) ([]model.Market, error) {
	var out []model.Market
	for _, id := range ids {
		if x, ok := m.markets[id]; ok {
			out = append(out, x)
		}
	}
	return out, nil
}

func (m *mem) ListByEvents(_ context.Context, ids []int64) ([]model.Market, error) {
	var out []model.Market
	for _, x := range m.markets {
		for _, id := range ids {
			if x.EventID == id {
				out = append(out, x)
			}
		}
	}
	return out, nil
}

func (m *mem) ListByEventForUpdate(ctx context.Context, id int64) ([]model.Market, error) {
	out, _ := m.ListByEvents(ctx, []int64{id})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *mem) ListRewarded(context.Context) ([]model.Market, error) {
	var out []model.Market
	for _, x := range m.markets {
		if x.Status == model.MarketOpen && x.RewardPool > 0 {
			out = append(out, x)
		}
	}
	return out, nil
}

func (m *mem) Search(context.Context, port.MarketFilter) ([]model.Market, int64, error) {
	return nil, 0, nil
}

func (m *mem) Save(_ context.Context, x *model.Market) error { m.markets[x.ID] = *x; return nil }

// ---- EventRepository ----

type eventRepo struct{ *mem }

func (r eventRepo) Create(_ context.Context, e *model.Event) error {
	e.ID = r.id()
	r.events[e.ID] = *e
	return nil
}
func (r eventRepo) GetByID(_ context.Context, id int64) (*model.Event, error) {
	e, ok := r.events[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &e, nil
}
func (r eventRepo) GetBySlug(context.Context, string) (*model.Event, error) {
	return nil, port.ErrNotFound
}
func (r eventRepo) GetByIDs(context.Context, []int64) ([]model.Event, error) { return nil, nil }
func (r eventRepo) List(context.Context, port.EventFilter) (*port.Page[model.Event], error) {
	return &port.Page[model.Event]{IsLastPage: true}, nil
}
func (r eventRepo) Related(context.Context, int64, string, int) ([]model.Event, error) {
	return nil, nil
}
func (r eventRepo) Categories(context.Context) ([]port.CategoryCount, error) { return nil, nil }
func (r eventRepo) AddVolume(_ context.Context, id, delta int64) error {
	e := r.events[id]
	e.Volume += delta
	r.events[id] = e
	return nil
}
func (r eventRepo) Save(_ context.Context, e *model.Event) error { r.events[e.ID] = *e; return nil }

// ---- OrderRepository ----

type orderRepo struct{ *mem }

func (r orderRepo) Create(_ context.Context, o *model.Order) error {
	o.ID = r.id()
	o.CreatedAt = r.stamp()
	r.orders[o.ID] = *o
	return nil
}
func (r orderRepo) GetByID(_ context.Context, id int64) (*model.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &o, nil
}
func (r orderRepo) GetByIDForUpdate(ctx context.Context, id int64) (*model.Order, error) {
	return r.GetByID(ctx, id)
}
func (r orderRepo) GetByClientID(_ context.Context, user, cid string) (*model.Order, error) {
	for _, o := range r.orders {
		if o.UserID == user && o.ClientOrderID == cid {
			o := o
			return &o, nil
		}
	}
	return nil, port.ErrNotFound
}
func (r orderRepo) Save(_ context.Context, o *model.Order) error { r.orders[o.ID] = *o; return nil }

func (r orderRepo) open(marketID int64, side model.BookSide) []model.Order {
	var out []model.Order
	for _, o := range r.orders {
		if o.MarketID == marketID && o.Status == model.OrderOpen && o.BookSide == side {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].YesPrice != out[j].YesPrice {
			if side == model.Bid {
				return out[i].YesPrice > out[j].YesPrice
			}
			return out[i].YesPrice < out[j].YesPrice
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (r orderRepo) ListCrossing(_ context.Context, marketID int64, side model.BookSide, yes int64, limit int) ([]model.Order, error) {
	var out []model.Order
	for _, o := range r.open(marketID, side) {
		if (side == model.Ask && o.YesPrice <= yes) || (side == model.Bid && o.YesPrice >= yes) {
			out = append(out, o)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (r orderRepo) ListOpenByMarketForUpdate(_ context.Context, marketID int64) ([]model.Order, error) {
	return append(r.open(marketID, model.Bid), r.open(marketID, model.Ask)...), nil
}
func (r orderRepo) ListOpenByMarket(_ context.Context, marketID int64) ([]model.Order, error) {
	return append(r.open(marketID, model.Bid), r.open(marketID, model.Ask)...), nil
}
func (r orderRepo) List(_ context.Context, f port.OrderFilter) (*port.Page[model.Order], error) {
	var out []model.Order
	for _, o := range r.orders {
		if (f.UserID == "" || o.UserID == f.UserID) && (f.MarketID == 0 || o.MarketID == f.MarketID) && (f.Status == "" || o.Status == f.Status) {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return &port.Page[model.Order]{Items: out, IsLastPage: true}, nil
}
func (r orderRepo) BookLevels(_ context.Context, marketID int64, side model.BookSide) ([]port.BookLevel, error) {
	var out []port.BookLevel
	for _, o := range r.open(marketID, side) {
		if n := len(out); n > 0 && out[n-1].YesPrice == o.YesPrice {
			out[n-1].Size += o.Remaining()
		} else {
			out = append(out, port.BookLevel{YesPrice: o.YesPrice, Size: o.Remaining()})
		}
	}
	return out, nil
}

// ---- TradeRepository ----

type tradeRepo struct{ *mem }

func (r tradeRepo) Create(_ context.Context, t *model.Trade) error {
	t.ID = r.id()
	t.CreatedAt = r.stamp()
	r.trades = append(r.trades, *t)
	return nil
}
func (r tradeRepo) ListByMarket(context.Context, int64, int, int) ([]model.Trade, error) {
	return r.trades, nil
}
func (r tradeRepo) ListByMarketSince(_ context.Context, id int64, _ time.Time, _ int) ([]model.Trade, error) {
	return r.trades, nil
}
func (r tradeRepo) ListByUser(_ context.Context, user string, _, _ int) ([]model.Trade, error) {
	var out []model.Trade
	for _, t := range r.trades {
		if t.MakerUserID == user || t.TakerUserID == user {
			out = append(out, t)
		}
	}
	return out, nil
}

// ---- BalanceRepository / LedgerRepository / PositionRepository ----

type balanceRepo struct{ *mem }

func (r balanceRepo) GetForUpdate(_ context.Context, user string) (*model.Balance, error) {
	b := r.balances[user]
	b.UserID = user
	return &b, nil
}
func (r balanceRepo) Get(ctx context.Context, user string) (*model.Balance, error) {
	return r.GetForUpdate(ctx, user)
}
func (r balanceRepo) Save(_ context.Context, b *model.Balance) error {
	r.balances[b.UserID] = *b
	return nil
}

type ledgerRepo struct{ *mem }

func (r ledgerRepo) Create(_ context.Context, e *model.LedgerEntry) error {
	e.ID = r.id()
	r.ledger = append(r.ledger, *e)
	return nil
}
func (r ledgerRepo) ListByUser(context.Context, string, int, int) ([]model.LedgerEntry, error) {
	return r.ledger, nil
}

type positionRepo struct{ *mem }

func (r positionRepo) GetForUpdate(_ context.Context, marketID int64, user string, o model.Outcome) (*model.Position, error) {
	p, ok := r.pos[posKey(marketID, user, o)]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &p, nil
}
func (r positionRepo) Save(_ context.Context, p *model.Position) error {
	if p.ID == 0 {
		p.ID = r.id()
	}
	r.pos[posKey(p.MarketID, p.UserID, p.Outcome)] = *p
	return nil
}
func (r positionRepo) ListByUser(_ context.Context, user string, settled bool) ([]model.Position, error) {
	var out []model.Position
	for _, p := range r.pos {
		if p.UserID == user && p.Settled == settled && (settled || p.Shares > 0) {
			out = append(out, p)
		}
	}
	return out, nil
}
func (r positionRepo) ListByMarketForUpdate(_ context.Context, marketID int64) ([]model.Position, error) {
	var out []model.Position
	for _, p := range r.pos {
		if p.MarketID == marketID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (r positionRepo) TopHolders(context.Context, int64, model.Outcome, int) ([]port.Holder, error) {
	return nil, nil
}

// ---- PnL / exchange account / rewards / referrals ----

type pnlRepo struct{ *mem }

func (r pnlRepo) CreateMany(_ context.Context, e []model.PnLEntry) error {
	r.pnl = append(r.pnl, e...)
	return nil
}

type exchangeRepo struct{ *mem }

func (r exchangeRepo) Add(_ context.Context, e ...model.ExchangeEntry) error {
	r.exch = append(r.exch, e...)
	return nil
}
func (r exchangeRepo) Balance(_ context.Context, b model.ExchangeBucket) (int64, error) {
	var sum int64
	for _, e := range r.exch {
		if e.Bucket == b {
			sum += e.Amount
		}
	}
	return sum, nil
}

type rewardRepo struct{ *mem }

func (r rewardRepo) ClaimEpoch(_ context.Context, start time.Time) (bool, error) {
	if r.epochs[start] {
		return false, nil
	}
	r.epochs[start] = true
	return true, nil
}
func (r rewardRepo) CreatePayouts(_ context.Context, p []model.RewardPayout) error {
	r.payouts = append(r.payouts, p...)
	return nil
}
func (r rewardRepo) ListPayouts(context.Context, string, int, string) (*port.Page[model.RewardPayout], error) {
	return &port.Page[model.RewardPayout]{Items: r.payouts, IsLastPage: true}, nil
}
func (r rewardRepo) TotalPaid(_ context.Context, user string) (int64, error) {
	var sum int64
	for _, p := range r.payouts {
		if p.UserID == user {
			sum += p.Amount
		}
	}
	return sum, nil
}

type referralRepo struct{ *mem }

func (r referralRepo) GetCode(_ context.Context, u string) (*model.ReferralCode, error) {
	c, ok := r.codes[u]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &c, nil
}
func (r referralRepo) SaveCode(_ context.Context, c *model.ReferralCode) error {
	r.codes[c.UserID] = *c
	return nil
}
func (r referralRepo) GetReferral(_ context.Context, u string) (*model.Referral, error) {
	x, ok := r.referrals[u]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &x, nil
}
func (r referralRepo) SaveReferral(_ context.Context, x *model.Referral) error {
	if _, ok := r.referrals[x.RefereeUserID]; ok {
		return port.ErrAlreadyExists
	}
	r.referrals[x.RefereeUserID] = *x
	return nil
}
func (r referralRepo) CountReferees(_ context.Context, u string) (int64, error) {
	var n int64
	for _, x := range r.referrals {
		if x.ReferrerUserID == u {
			n++
		}
	}
	return n, nil
}
func (r referralRepo) Earnings(_ context.Context, u string) (int64, error) {
	var sum int64
	for _, t := range r.trades {
		if t.ReferrerUserID == u {
			sum += t.ReferralFee
		}
	}
	return sum, nil
}

type fakeReferralGateway struct {
	owner    string
	err      error
	activate []string
}

func (g *fakeReferralGateway) GenerateLink(context.Context, string) (*port.ReferralLink, error) {
	return &port.ReferralLink{RefCode: "CODE123", DeepLink: "app://open?ref=CODE123"}, g.err
}
func (g *fakeReferralGateway) Activate(_ context.Context, code, user string) (string, error) {
	g.activate = append(g.activate, code+":"+user)
	return g.owner, g.err
}
func (g *fakeReferralGateway) Stats(context.Context, string) (*port.ReferralStats, error) {
	return &port.ReferralStats{TotalInvited: 3}, nil
}

// ---- gateways ----

type fakeWallets struct {
	wallet      port.Wallet
	transferErr error
	transfers   []port.TransferInput
	reversed    []string
}

func (w *fakeWallets) GetWalletByUser(context.Context, string) (*port.Wallet, error) {
	c := w.wallet
	return &c, nil
}
func (w *fakeWallets) Transfer(_ context.Context, in port.TransferInput) (string, error) {
	if w.transferErr != nil {
		return "", w.transferErr
	}
	w.transfers = append(w.transfers, in)
	return fmt.Sprintf("tr-%d", len(w.transfers)), nil
}
func (w *fakeWallets) ReverseTransfer(_ context.Context, id, _ string) error {
	w.reversed = append(w.reversed, id)
	return nil
}

type fakeNotifier struct{ sent []port.Notification }

func (n *fakeNotifier) Notify(_ context.Context, m port.Notification) error {
	n.sent = append(n.sent, m)
	return nil
}

type fakePublisher struct{ msgs []port.StreamMessage }

func (p *fakePublisher) Publish(_ string, payload any) {
	if m, ok := payload.(port.StreamMessage); ok {
		p.msgs = append(p.msgs, m)
	}
}

// ---- assembled environment ----

const treasury = "wallet-treasury"

type env struct {
	mem        *mem
	wallets    *fakeWallets
	notifier   *fakeNotifier
	pub        *fakePublisher
	referrals  *fakeReferralGateway
	exchange   *exchangeService
	resolution *resolutionService
	accounts   *accountService
	rewards    *rewardService
	referral   *referralService
	market     model.Market
	deposited  int64
}

func newEnv() *env { return newEnvWith(Options{}) }

// newEnvWith builds the whole application on the in-memory stores. Options
// left zero get the test defaults: USD, share value 100, no fees.
func newEnvWith(opts Options) *env {
	m := newMem()
	wallets := &fakeWallets{wallet: port.Wallet{ID: "wallet-user", Currency: "USD", Balance: 1_000_000, Status: "active"}}
	notifier := &fakeNotifier{}
	pub := &fakePublisher{}
	gateway := &fakeReferralGateway{owner: "ref-owner"}
	opts.TreasuryWalletID = treasury
	opts.Currency, opts.ShareValue = "USD", 100
	if opts.DisputeWindow == 0 {
		opts.DisputeWindow = time.Hour
	}

	e := &env{mem: m, wallets: wallets, notifier: notifier, pub: pub, referrals: gateway}
	e.exchange = NewExchangeService(ExchangeParams{
		Markets: m, Events: eventRepo{m}, Orders: orderRepo{m}, Trades: tradeRepo{m}, Balances: balanceRepo{m},
		Positions: positionRepo{m}, Ledger: ledgerRepo{m}, Referrals: referralRepo{m}, PnL: pnlRepo{m},
		Exchange: exchangeRepo{m}, Tx: m, Notifier: notifier, Publisher: pub, Opts: opts,
	})
	e.resolution = NewResolutionService(ResolutionParams{
		Markets: m, Events: eventRepo{m}, Orders: orderRepo{m}, Positions: positionRepo{m}, Balances: balanceRepo{m},
		Ledger: ledgerRepo{m}, Referrals: referralRepo{m}, PnL: pnlRepo{m}, Exchange: exchangeRepo{m},
		Tx: m, Notifier: notifier, Publisher: pub, Opts: opts,
	})
	e.accounts = NewAccountService(AccountParams{
		Markets: m, Trades: tradeRepo{m}, Balances: balanceRepo{m}, Ledger: ledgerRepo{m}, Positions: positionRepo{m},
		Exchange: exchangeRepo{m}, Tx: m, Wallets: wallets, Notifier: notifier, Opts: opts,
	})
	e.rewards = NewRewardService(RewardParams{
		Markets: m, Orders: orderRepo{m}, Positions: positionRepo{m}, Balances: balanceRepo{m}, Referrals: referralRepo{m},
		PnL: pnlRepo{m}, Exchange: exchangeRepo{m}, Rewards: rewardRepo{m}, Tx: m, Notifier: notifier, Opts: opts,
	})
	e.referral = NewReferralService(ReferralParams{
		Referrals: referralRepo{m}, Gateway: gateway, Trades: tradeRepo{m}, Notifier: notifier, Opts: opts,
	})

	catalog := NewCatalogService(eventRepo{m}, m, m, opts)
	ev, err := catalog.CreateEvent(context.Background(), port.CreateEventInput{
		Title: "Will it rain?", Category: "weather", EndDate: time.Now().Add(time.Hour),
		Markets: []port.CreateMarketInput{{Question: "Will it rain tomorrow?"}},
	})
	if err != nil {
		panic(err)
	}
	e.market = ev.Markets[0]
	return e
}

// fund deposits cash for a user through the real deposit flow.
func (e *env) fund(user string, amount int64) {
	if _, err := e.accounts.Deposit(context.Background(), user, amount); err != nil {
		panic(err)
	}
	e.deposited += amount
}
