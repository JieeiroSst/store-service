package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

type ResolutionParams struct {
	fx.In

	Markets   port.MarketRepository
	Events    port.EventRepository
	Orders    port.OrderRepository
	Positions port.PositionRepository
	Balances  port.BalanceRepository
	Ledger    port.LedgerRepository
	Referrals port.ReferralRepository
	PnL       port.PnLRepository
	Exchange  port.ExchangeAccountRepository
	Tx        port.TxManager
	Notifier  port.Notifier
	Publisher port.Publisher
	Opts      Options
}

type resolutionService struct {
	markets   port.MarketRepository
	events    port.EventRepository
	orders    port.OrderRepository
	positions port.PositionRepository
	balances  port.BalanceRepository
	ledger    port.LedgerRepository
	tx        port.TxManager
	notifier  port.Notifier
	publisher port.Publisher
	opts      Options
	deps      sessionDeps
	now       func() time.Time
}

func NewResolutionService(p ResolutionParams) *resolutionService {
	return &resolutionService{
		markets: p.Markets, events: p.Events, orders: p.Orders, positions: p.Positions, balances: p.Balances,
		ledger: p.Ledger, tx: p.Tx, notifier: p.Notifier, publisher: p.Publisher, opts: p.Opts.withDefaults(), now: time.Now,
		deps: sessionDeps{
			balances: p.Balances, positions: p.Positions, referrals: p.Referrals,
			pnl: p.PnL, exchange: p.Exchange, orders: p.Orders,
		},
	}
}

func validResolution(o model.Outcome) bool {
	return o == model.OutcomeYes || o == model.OutcomeNo || o == model.OutcomeSplit
}

func errBadOutcome() error {
	return fmt.Errorf("%w: outcome must be yes, no or split", port.ErrInvalidInput)
}

type group struct {
	event   *model.Event
	markets []model.Market
}

func (s *resolutionService) lockGroup(ctx context.Context, marketID int64) (*group, error) {
	peek, err := s.markets.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}
	event, err := s.events.GetByID(ctx, peek.EventID)
	if err != nil {
		return nil, err
	}
	if event.NegRisk {
		ms, err := s.markets.ListByEventForUpdate(ctx, event.ID)
		return &group{event: event, markets: ms}, err
	}
	m, err := s.markets.GetByIDForUpdate(ctx, marketID)
	if err != nil {
		return nil, err
	}
	return &group{event: event, markets: []model.Market{*m}}, nil
}

func (s *resolutionService) lockEvent(ctx context.Context, eventID int64) (*group, error) {
	event, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if !event.NegRisk {
		return nil, port.ErrNotNegRisk
	}
	ms, err := s.markets.ListByEventForUpdate(ctx, eventID)
	return &group{event: event, markets: ms}, err
}

func (g *group) find(id int64) *model.Market {
	for i := range g.markets {
		if g.markets[i].ID == id {
			return &g.markets[i]
		}
	}
	return nil
}

func (g *group) save(ctx context.Context, repo port.MarketRepository) error {
	for i := range g.markets {
		if err := repo.Save(ctx, &g.markets[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *resolutionService) publishStatus(g *group, requested int64) *model.Market {
	var out *model.Market
	for i := range g.markets {
		m := &g.markets[i]
		s.publisher.Publish(port.MarketTopic(m.ID), port.StreamMessage{Type: "status", Payload: m})
		if m.ID == requested {
			out = m
		}
	}
	return out
}

func (s *resolutionService) Propose(ctx context.Context, marketID int64, outcome model.Outcome) (*model.Market, error) {
	if !validResolution(outcome) {
		return nil, errBadOutcome()
	}
	var g *group
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if g, err = s.lockGroup(ctx, marketID); err != nil {
			return err
		}
		if g.event.NegRisk {
			return port.ErrNegRisk
		}
		m := &g.markets[0]
		if m.Status != model.MarketOpen {
			return port.ErrInvalidTransition
		}
		s.openWindow(m, outcome)
		return g.save(ctx, s.markets)
	})
	if err != nil {
		return nil, err
	}
	return s.publishStatus(g, marketID), nil
}

func (s *resolutionService) openWindow(m *model.Market, outcome model.Outcome) {
	deadline := s.now().Add(s.opts.DisputeWindow)
	m.Status = model.MarketProposed
	m.ProposedOutcome = outcome
	m.DisputeDeadline = &deadline
}

func (s *resolutionService) ProposeEvent(ctx context.Context, eventID, winnerMarketID int64) (*model.Event, error) {
	var g *group
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if g, err = s.lockEvent(ctx, eventID); err != nil {
			return err
		}
		if g.find(winnerMarketID) == nil {
			return fmt.Errorf("%w: winner_market_id must belong to the event", port.ErrInvalidInput)
		}
		for i := range g.markets {
			m := &g.markets[i]
			if m.Status != model.MarketOpen {
				return port.ErrInvalidTransition
			}
			outcome := model.OutcomeNo
			if m.ID == winnerMarketID {
				outcome = model.OutcomeYes
			}
			s.openWindow(m, outcome)
		}
		return g.save(ctx, s.markets)
	})
	if err != nil {
		return nil, err
	}
	s.publishStatus(g, 0)
	g.event.Markets = g.markets
	return g.event, nil
}

func (s *resolutionService) Dispute(ctx context.Context, marketID int64, userID, reason string) (*model.Market, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	var g *group
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if g, err = s.lockGroup(ctx, marketID); err != nil {
			return err
		}
		for i := range g.markets {
			m := &g.markets[i]
			if m.Status != model.MarketProposed {
				return port.ErrInvalidTransition
			}
			if m.DisputeDeadline == nil || !s.now().Before(*m.DisputeDeadline) {
				return port.ErrDisputeWindowShut
			}
		}

		sess := newSession(s.deps)
		target := g.find(marketID)
		if bond := s.opts.DisputeBond; bond > 0 {
			bal, err := sess.balance(ctx, userID)
			if err != nil {
				return err
			}
			if bal.Available < bond {
				return port.ErrInsufficientBalance
			}
			bal.Available -= bond
			target.DisputeBond = bond
			sess.addExchange(model.BucketBond, "dispute_bond", bond, target.ID, userID)
			if err := s.ledger.Create(ctx, &model.LedgerEntry{UserID: userID, Type: model.LedgerBond, Amount: -bond, MarketID: target.ID}); err != nil {
				return err
			}
		}
		for i := range g.markets {
			m := &g.markets[i]
			m.Status = model.MarketDisputed
			m.DisputedBy = userID
			m.DisputeReason = reason
		}
		if err := sess.flush(ctx); err != nil {
			return err
		}
		return g.save(ctx, s.markets)
	})
	if err != nil {
		return nil, err
	}
	return s.publishStatus(g, marketID), nil
}

func (s *resolutionService) Finalize(ctx context.Context, marketID int64) (*model.Market, error) {
	return s.settle(ctx, func(ctx context.Context) (*group, error) { return s.lockGroup(ctx, marketID) },
		marketID, func(g *group) (map[int64]model.Outcome, bool, error) {
			out := map[int64]model.Outcome{}
			for i := range g.markets {
				m := &g.markets[i]
				if m.Status != model.MarketProposed {
					return nil, false, port.ErrInvalidTransition
				}
				if m.DisputeDeadline != nil && s.now().Before(*m.DisputeDeadline) {
					return nil, false, port.ErrDisputeWindowOpen
				}
				out[m.ID] = m.ProposedOutcome
			}
			return out, false, nil
		})
}

func (s *resolutionService) Resolve(ctx context.Context, marketID int64, outcome model.Outcome) (*model.Market, error) {
	if !validResolution(outcome) {
		return nil, errBadOutcome()
	}
	return s.settle(ctx, func(ctx context.Context) (*group, error) { return s.lockGroup(ctx, marketID) },
		marketID, func(g *group) (map[int64]model.Outcome, bool, error) {
			if g.event.NegRisk {
				return nil, false, port.ErrNegRisk
			}
			if g.markets[0].Status == model.MarketResolved {
				return nil, false, port.ErrInvalidTransition
			}
			return map[int64]model.Outcome{g.markets[0].ID: outcome}, false, nil
		})
}

func (s *resolutionService) ResolveEvent(ctx context.Context, eventID, winnerMarketID int64) (*model.Event, error) {
	m, err := s.settle(ctx, func(ctx context.Context) (*group, error) { return s.lockEvent(ctx, eventID) },
		0, func(g *group) (map[int64]model.Outcome, bool, error) {
			if g.find(winnerMarketID) == nil {
				return nil, false, fmt.Errorf("%w: winner_market_id must belong to the event", port.ErrInvalidInput)
			}
			out := map[int64]model.Outcome{}
			for i := range g.markets {
				if g.markets[i].Status == model.MarketResolved {
					return nil, false, port.ErrInvalidTransition
				}
				out[g.markets[i].ID] = model.OutcomeNo
			}
			out[winnerMarketID] = model.OutcomeYes
			return out, true, nil
		})
	if err != nil {
		return nil, err
	}
	event, err := s.events.GetByID(ctx, m.EventID)
	if err != nil {
		return nil, err
	}
	ms, err := s.markets.ListByEvents(ctx, []int64{eventID})
	event.Markets = ms
	return event, err
}

type payout struct {
	market *model.Market
	userID string
	amount int64
}

func (s *resolutionService) settle(
	ctx context.Context,
	lock func(context.Context) (*group, error),
	requested int64,
	decide func(*group) (outcomes map[int64]model.Outcome, isEvent bool, err error),
) (*model.Market, error) {
	var (
		g       *group
		payouts []payout
	)
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		payouts = nil
		var err error
		if g, err = lock(ctx); err != nil {
			return err
		}
		outcomes, _, err := decide(g)
		if err != nil {
			return err
		}
		upheld := s.overturned(g, outcomes)
		for i := range g.markets {
			ps, err := s.settleOne(ctx, &g.markets[i], outcomes[g.markets[i].ID], upheld)
			if err != nil {
				return err
			}
			payouts = append(payouts, ps...)
		}
		return s.closeEventIfDone(ctx, g.event.ID)
	})
	if err != nil {
		return nil, err
	}

	bg := context.WithoutCancel(ctx)
	var out *model.Market
	for i := range g.markets {
		m := &g.markets[i]
		s.publisher.Publish(port.MarketTopic(m.ID), port.StreamMessage{Type: "resolved", Payload: m})
		if m.ID == requested || (requested == 0 && out == nil) {
			out = m
		}
	}
	go func() {
		for _, p := range payouts {
			notify(bg, s.notifier, port.Notification{
				UserID: p.userID, Title: "Market resolved",
				Message: fmt.Sprintf("\"%s\" resolved %s. %d was credited to your balance.", p.market.Question, strings.ToUpper(string(p.market.ResolvedOutcome)), p.amount),
			})
		}
	}()
	return out, nil
}

func (s *resolutionService) overturned(g *group, outcomes map[int64]model.Outcome) bool {
	for i := range g.markets {
		m := &g.markets[i]
		if m.ProposedOutcome != "" && m.ProposedOutcome != outcomes[m.ID] {
			return true
		}
	}
	return false
}

func (s *resolutionService) settleOne(ctx context.Context, m *model.Market, outcome model.Outcome, disputeUpheld bool) ([]payout, error) {
	open, err := s.orders.ListOpenByMarketForUpdate(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	sess := newSession(s.deps)
	for i := range open {
		if err := closeOrder(ctx, sess, &open[i], model.OrderCanceled); err != nil {
			return nil, err
		}
	}
	if err := sess.flush(ctx); err != nil {
		return nil, err
	}

	positions, err := s.positions.ListByMarketForUpdate(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	paid := newSession(s.deps)
	totals := map[string]int64{}
	for i := range positions {
		p := &positions[i]
		if p.Settled {
			continue
		}
		amount := settlementValue(p, outcome, m.ShareValue)
		p.RealizedPnL += amount - p.CostBasis
		paid.addPnL(p.UserID, m.ID, model.PnLSettlement, amount-p.CostBasis)
		p.CostBasis, p.Shares, p.LockedShares, p.Settled = 0, 0, 0, true
		if err := s.positions.Save(ctx, p); err != nil {
			return nil, err
		}
		if amount > 0 {
			bal, err := paid.balance(ctx, p.UserID)
			if err != nil {
				return nil, err
			}
			bal.Available += amount
			totals[p.UserID] += amount
		}
	}

	if m.DisputeBond > 0 && !m.BondSettled {
		if err := s.settleBond(ctx, paid, m, disputeUpheld); err != nil {
			return nil, err
		}
	}
	if err := paid.flush(ctx); err != nil {
		return nil, err
	}

	var out []payout
	for user, amount := range totals {
		if err := s.ledger.Create(ctx, &model.LedgerEntry{UserID: user, Type: model.LedgerSettlement, Amount: amount, MarketID: m.ID}); err != nil {
			return nil, err
		}
		out = append(out, payout{market: m, userID: user, amount: amount})
	}

	m.Status = model.MarketResolved
	m.ResolvedOutcome = outcome
	m.BestBid, m.BestAsk = 0, 0
	if err := s.markets.Save(ctx, m); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *resolutionService) settleBond(ctx context.Context, sess *session, m *model.Market, upheld bool) error {
	bond := m.DisputeBond
	m.BondSettled = true
	sess.addExchange(model.BucketBond, "dispute_bond_release", -bond, m.ID, m.DisputedBy)
	if upheld {
		bal, err := sess.balance(ctx, m.DisputedBy)
		if err != nil {
			return err
		}
		bal.Available += bond
		return s.ledger.Create(ctx, &model.LedgerEntry{UserID: m.DisputedBy, Type: model.LedgerBondRefund, Amount: bond, MarketID: m.ID})
	}
	sess.addExchange(model.BucketRevenue, "dispute_bond_forfeit", bond, m.ID, m.DisputedBy)
	sess.addPnL(m.DisputedBy, m.ID, model.PnLBond, -bond)
	return nil
}

func settlementValue(p *model.Position, outcome model.Outcome, shareValue int64) int64 {
	switch {
	case outcome == model.OutcomeSplit:
		return p.Shares * shareValue / 2
	case p.Outcome == outcome:
		return p.Shares * shareValue
	default:
		return 0
	}
}

func (s *resolutionService) closeEventIfDone(ctx context.Context, eventID int64) error {
	markets, err := s.markets.ListByEvents(ctx, []int64{eventID})
	if err != nil {
		return err
	}
	for _, m := range markets {
		if m.Status != model.MarketResolved {
			return nil
		}
	}
	event, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return err
	}
	event.Status = model.EventResolved
	return s.events.Save(ctx, event)
}
