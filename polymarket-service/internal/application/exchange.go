package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

func logf(format string, args ...any) { log.Printf(format, args...) }

func notify(ctx context.Context, n port.Notifier, msg port.Notification) {
	if err := n.Notify(ctx, msg); err != nil {
		logf("notify user %s: %v", msg.UserID, err)
	}
}

const makerBatch = 50

type ExchangeParams struct {
	fx.In

	Markets   port.MarketRepository
	Events    port.EventRepository
	Orders    port.OrderRepository
	Trades    port.TradeRepository
	Balances  port.BalanceRepository
	Positions port.PositionRepository
	Ledger    port.LedgerRepository
	Referrals port.ReferralRepository
	PnL       port.PnLRepository
	Exchange  port.ExchangeAccountRepository
	Tx        port.TxManager
	Notifier  port.Notifier
	Publisher port.Publisher
	Opts      Options
}

type exchangeService struct {
	markets   port.MarketRepository
	events    port.EventRepository
	orders    port.OrderRepository
	trades    port.TradeRepository
	balances  port.BalanceRepository
	positions port.PositionRepository
	ledger    port.LedgerRepository
	tx        port.TxManager
	notifier  port.Notifier
	publisher port.Publisher
	opts      Options
	deps      sessionDeps
	now       func() time.Time
}

func NewExchangeService(p ExchangeParams) *exchangeService {
	return &exchangeService{
		markets: p.Markets, events: p.Events, orders: p.Orders, trades: p.Trades,
		balances: p.Balances, positions: p.Positions, ledger: p.Ledger, tx: p.Tx,
		notifier: p.Notifier, publisher: p.Publisher, opts: p.Opts.withDefaults(), now: time.Now,
		deps: sessionDeps{
			balances: p.Balances, positions: p.Positions, referrals: p.Referrals,
			pnl: p.PnL, exchange: p.Exchange, orders: p.Orders,
		},
	}
}

func (s *exchangeService) session() *session { return newSession(s.deps) }

// takerFee is the fee on cash the taker spends or receives. It rounds down, so
// the fee on a whole order never exceeds the fee reserved for it.
func (s *exchangeService) takerFee(cash int64) int64 { return cash * s.opts.TakerFeeBps / 10_000 }

// PlaceOrder validates, escrows, matches and (for resting orders) books an
// order in one transaction. All orders in a market serialise on the market row
// lock, so price-time priority is exactly id order.
func (s *exchangeService) PlaceOrder(ctx context.Context, in port.PlaceOrderInput) (*port.PlaceOrderResult, error) {
	if in.UserID == "" {
		return nil, port.ErrInvalidUser
	}
	if !in.Outcome.Valid() {
		return nil, port.ErrInvalidOutcome
	}
	if in.Side != model.Buy && in.Side != model.Sell {
		return nil, fmt.Errorf("%w: side must be buy or sell", port.ErrInvalidInput)
	}
	if in.Type == "" {
		in.Type = model.LimitOrder
	}

	var result *port.PlaceOrderResult
	var market *model.Market
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		m, err := s.markets.GetByIDForUpdate(ctx, in.MarketID)
		if err != nil {
			return err
		}
		market = m
		if !m.Tradable(s.now()) {
			return port.ErrMarketNotTradable
		}
		if in.ClientOrderID != "" {
			existing, err := s.orders.GetByClientID(ctx, in.UserID, in.ClientOrderID)
			if err == nil {
				result = &port.PlaceOrderResult{Order: existing}
				return nil
			}
			if !errors.Is(err, port.ErrNotFound) {
				return err
			}
		}

		order, err := s.buildOrder(m, in)
		if err != nil {
			return err
		}
		sess := s.session()
		if err := s.escrow(ctx, sess, order); err != nil {
			return err
		}
		if err := s.orders.Create(ctx, order); err != nil {
			return err
		}

		fills, err := s.match(ctx, sess, m, order)
		if err != nil {
			return err
		}
		if err := s.finish(ctx, sess, order); err != nil {
			return err
		}
		if err := sess.flush(ctx); err != nil {
			return err
		}
		if err := s.orders.Save(ctx, order); err != nil {
			return err
		}
		if err := s.refreshMarket(ctx, m, fills); err != nil {
			return err
		}
		result = &port.PlaceOrderResult{Order: order, Trades: fills}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.afterOrder(ctx, market, result)
	return result, nil
}

func (s *exchangeService) buildOrder(m *model.Market, in port.PlaceOrderInput) (*model.Order, error) {
	v := m.ShareValue
	o := &model.Order{
		MarketID: m.ID, UserID: in.UserID, ClientOrderID: in.ClientOrderID,
		Outcome: in.Outcome, Side: in.Side, Type: in.Type, Status: model.OrderOpen,
	}

	switch in.Type {
	case model.LimitOrder:
		o.TimeInForce = in.TimeInForce
		if o.TimeInForce == "" {
			o.TimeInForce = model.GTC
			if in.ExpiresAt != nil {
				o.TimeInForce = model.GTD
			}
		}
		if in.Price < 1 || in.Price > v-1 {
			return nil, port.ErrInvalidPrice
		}
		if in.Size < m.MinOrderSize {
			return nil, port.ErrInvalidSize
		}
		o.Price, o.Size = in.Price, in.Size
		if o.Side == model.Buy {
			o.LockedCash = o.Price * o.Size
			// Reserve the worst-case taker fee. It is released as soon as the
			// order stops being a taker (it rests, or finishes).
			o.FeeReserve = s.takerFee(o.LockedCash)
		}
	case model.MarketOrder:
		o.TimeInForce = in.TimeInForce
		if o.TimeInForce == "" {
			o.TimeInForce = model.FAK
		}
		if o.Side == model.Buy {
			// A market buy spends a cash budget, fee included, on whatever the
			// book offers.
			if in.Amount <= 0 {
				return nil, fmt.Errorf("%w: market buy needs a positive amount", port.ErrInvalidInput)
			}
			if o.TimeInForce != model.FAK {
				return nil, fmt.Errorf("%w: market buys are fill-and-kill", port.ErrInvalidInput)
			}
			o.Price, o.Budget, o.Size, o.LockedCash = v-1, in.Amount, in.Amount, in.Amount
		} else {
			if in.Size < 1 {
				return nil, port.ErrInvalidSize
			}
			o.Price, o.Size = 1, in.Size
		}
	default:
		return nil, fmt.Errorf("%w: type must be limit or market", port.ErrInvalidInput)
	}

	switch o.TimeInForce {
	case model.GTC, model.FOK, model.FAK:
	case model.GTD:
		if in.ExpiresAt == nil || !in.ExpiresAt.After(s.now()) {
			return nil, fmt.Errorf("%w: gtd orders need a future expires_at", port.ErrInvalidInput)
		}
		o.ExpiresAt = in.ExpiresAt
	default:
		return nil, fmt.Errorf("%w: unknown time_in_force", port.ErrInvalidInput)
	}
	if o.Type == model.MarketOrder && (o.TimeInForce == model.GTC || o.TimeInForce == model.GTD) {
		return nil, fmt.Errorf("%w: market orders cannot rest on the book", port.ErrInvalidInput)
	}

	o.BookSide, o.YesPrice = model.BookPlacement(o.Outcome, o.Side, o.Price, v)
	return o, nil
}

// escrow reserves what the order needs: cash for buys, shares for sells.
func (s *exchangeService) escrow(ctx context.Context, sess *session, o *model.Order) error {
	if o.Side == model.Buy {
		bal, err := sess.balance(ctx, o.UserID)
		if err != nil {
			return err
		}
		need := o.LockedCash + o.FeeReserve
		if bal.Available < need {
			return port.ErrInsufficientBalance
		}
		bal.Available -= need
		bal.Locked += need
		return nil
	}
	pos, err := sess.position(ctx, o.MarketID, o.UserID, o.Outcome)
	if err != nil {
		return err
	}
	if pos.Free() < o.Size {
		return port.ErrInsufficientShares
	}
	pos.LockedShares += o.Size
	return nil
}

// match runs the taker against the opposite side of the book, best price
// first, executing every trade at the resting order's price.
func (s *exchangeService) match(ctx context.Context, sess *session, m *model.Market, taker *model.Order) ([]model.Trade, error) {
	var fills []model.Trade
	makerSide := model.Ask
	if taker.BookSide == model.Ask {
		makerSide = model.Bid
	}

	for s.hasCapacity(taker) {
		makers, err := s.orders.ListCrossing(ctx, m.ID, makerSide, taker.YesPrice, makerBatch)
		if err != nil {
			return nil, err
		}
		if len(makers) == 0 {
			break
		}
		stop := false
		for i := range makers {
			maker := &makers[i]
			if !s.hasCapacity(taker) {
				break
			}
			switch {
			case maker.ExpiresAt != nil && !s.now().Before(*maker.ExpiresAt):
				err = closeOrder(ctx, sess, maker, model.OrderExpired)
			case maker.UserID == taker.UserID:
				// Self-trade prevention: the resting order is cancelled.
				err = closeOrder(ctx, sess, maker, model.OrderCanceled)
			default:
				var fill *model.Trade
				if fill, err = s.execute(ctx, sess, m, taker, maker); fill != nil {
					fills = append(fills, *fill)
				} else if err == nil {
					stop = true // the taker cannot afford even one more share
				}
			}
			if err != nil {
				return nil, err
			}
			if stop {
				break
			}
		}
		if stop {
			break
		}
	}
	return fills, nil
}

func (s *exchangeService) hasCapacity(o *model.Order) bool {
	if o.IsBudget() {
		return o.LockedCash > 0
	}
	return o.Remaining() > 0
}

// budgetShares is how many shares at unit price a market buy can afford with
// fee included.
func (s *exchangeService) budgetShares(budget, unit int64) int64 {
	if budget <= 0 {
		return 0
	}
	n := budget * 10_000 / (unit * (10_000 + s.opts.TakerFeeBps))
	for n > 0 && unit*n+s.takerFee(unit*n) > budget {
		n--
	}
	for unit*(n+1)+s.takerFee(unit*(n+1)) <= budget {
		n++
	}
	return n
}

// execute trades the taker against one resting order at the maker's price. It
// returns nil when the taker's remaining budget cannot buy a single share.
func (s *exchangeService) execute(ctx context.Context, sess *session, m *model.Market, taker, maker *model.Order) (*model.Trade, error) {
	price := maker.YesPrice
	unit := model.OwnPrice(taker.Outcome, price, m.ShareValue)
	n := min(taker.Remaining(), maker.Remaining())
	if taker.IsBudget() {
		n = min(n, s.budgetShares(taker.LockedCash, unit))
	}
	if n <= 0 {
		return nil, nil
	}
	fee := s.takerFee(unit * n)

	if err := s.applyFill(ctx, sess, m, taker, price, n, fee); err != nil {
		return nil, err
	}
	if err := s.applyFill(ctx, sess, m, maker, price, n, 0); err != nil {
		return nil, err
	}
	if maker.Remaining() == 0 {
		maker.Status = model.OrderFilled
	}
	if err := s.orders.Save(ctx, maker); err != nil {
		return nil, err
	}

	kind := model.KindSwap
	switch {
	case taker.Side == model.Buy && maker.Side == model.Buy:
		kind = model.KindMint
	case taker.Side == model.Sell && maker.Side == model.Sell:
		kind = model.KindMerge
	}
	trade := &model.Trade{
		MarketID: m.ID, YesPrice: price, Size: n, Kind: kind,
		MakerOrderID: maker.ID, TakerOrderID: taker.ID,
		MakerUserID: maker.UserID, TakerUserID: taker.UserID,
		MakerOutcome: maker.Outcome, MakerSide: maker.Side,
		TakerOutcome: taker.Outcome, TakerSide: taker.Side,
		TakerFee: fee,
	}
	if fee > 0 {
		if err := s.splitFee(ctx, sess, m, trade); err != nil {
			return nil, err
		}
	}
	if err := s.trades.Create(ctx, trade); err != nil {
		return nil, err
	}
	return trade, nil
}

// applyFill settles n shares of one party's order at a YES-terms price. The
// taker fee, if any, comes out of what the taker spends or receives.
func (s *exchangeService) applyFill(ctx context.Context, sess *session, m *model.Market, o *model.Order, yesPrice, n, fee int64) error {
	own := model.OwnPrice(o.Outcome, yesPrice, m.ShareValue)
	cash := own * n

	bal, err := sess.balance(ctx, o.UserID)
	if err != nil {
		return err
	}
	pos, err := sess.position(ctx, o.MarketID, o.UserID, o.Outcome)
	if err != nil {
		return err
	}

	if o.Side == model.Buy {
		// Limit buys reserved their limit price, so a better fill hands the
		// difference back. Budget buys reserved exactly cash plus fee.
		if o.IsBudget() {
			bal.Locked -= cash + fee
			o.LockedCash -= cash + fee
		} else {
			bal.Locked -= o.Price * n
			bal.Available += o.Price*n - cash
			o.LockedCash -= o.Price * n

			fromReserve := min(fee, o.FeeReserve)
			bal.Locked -= fromReserve
			o.FeeReserve -= fromReserve
			bal.Available -= fee - fromReserve
		}
		pos.Shares += n
		pos.CostBasis += cash
	} else {
		released := pos.CostBasis * n / pos.Shares
		pos.CostBasis -= released
		pos.RealizedPnL += cash - released
		sess.addPnL(o.UserID, o.MarketID, model.PnLTrade, cash-released)
		pos.Shares -= n
		pos.LockedShares -= n
		bal.Available += cash - fee
	}
	o.Filled += n
	o.FilledCash += cash
	o.Fee += fee
	sess.addPnL(o.UserID, o.MarketID, model.PnLFee, -fee)
	return nil
}

// splitFee pays out a taker fee that has already been collected: a rebate to
// the maker, a commission to the taker's referrer, and the rest to the
// exchange. Rounding always favours the exchange.
func (s *exchangeService) splitFee(ctx context.Context, sess *session, m *model.Market, t *model.Trade) error {
	referrer, err := sess.referrer(ctx, t.TakerUserID)
	if err != nil {
		return err
	}
	t.MakerRebate = t.TakerFee * s.opts.MakerRebateBps / 10_000
	if referrer != "" && referrer != t.TakerUserID {
		t.ReferralFee = t.TakerFee * s.opts.ReferralBps / 10_000
		t.ReferrerUserID = referrer
	}

	if t.MakerRebate > 0 {
		bal, err := sess.balance(ctx, t.MakerUserID)
		if err != nil {
			return err
		}
		bal.Available += t.MakerRebate
		sess.addPnL(t.MakerUserID, m.ID, model.PnLRebate, t.MakerRebate)
	}
	if t.ReferralFee > 0 {
		bal, err := sess.balance(ctx, referrer)
		if err != nil {
			return err
		}
		bal.Available += t.ReferralFee
		sess.addPnL(referrer, m.ID, model.PnLReferral, t.ReferralFee)
	}
	sess.addExchange(model.BucketRevenue, "fee", t.TakerFee-t.MakerRebate-t.ReferralFee, m.ID, "")
	return nil
}

func (s *exchangeService) finish(ctx context.Context, sess *session, o *model.Order) error {
	if o.FeeReserve > 0 {
		bal, err := sess.balance(ctx, o.UserID)
		if err != nil {
			return err
		}
		bal.Locked -= o.FeeReserve
		bal.Available += o.FeeReserve
		o.FeeReserve = 0
	}

	fullyFilled := o.Remaining() == 0
	if o.IsBudget() {
		fullyFilled = false
		o.Size = o.Filled
	}

	switch {
	case fullyFilled:
		o.Status = model.OrderFilled
		return nil
	case o.TimeInForce == model.FOK:
		return port.ErrNotFilled
	case o.TimeInForce == model.FAK:
		if o.Filled == 0 {
			return closeOrder(ctx, sess, o, model.OrderCanceled)
		}
		status := model.OrderCanceled
		if o.Remaining() == 0 {
			status = model.OrderFilled
		}
		return closeOrder(ctx, sess, o, status)
	}
	return nil
}

func (s *exchangeService) refreshMarket(ctx context.Context, m *model.Market, fills []model.Trade) error {
	bids, err := s.orders.BookLevels(ctx, m.ID, model.Bid)
	if err != nil {
		return err
	}
	asks, err := s.orders.BookLevels(ctx, m.ID, model.Ask)
	if err != nil {
		return err
	}
	m.BestBid, m.BestAsk = 0, 0
	if len(bids) > 0 {
		m.BestBid = bids[0].YesPrice
	}
	if len(asks) > 0 {
		m.BestAsk = asks[0].YesPrice
	}

	var volume int64
	for _, f := range fills {
		volume += f.YesPrice * f.Size
		m.LastPrice = f.YesPrice
	}
	m.Volume += volume
	if err := s.markets.Save(ctx, m); err != nil {
		return err
	}
	if volume > 0 {
		return s.events.AddVolume(ctx, m.EventID, volume)
	}
	return nil
}

func (s *exchangeService) afterOrder(ctx context.Context, m *model.Market, r *port.PlaceOrderResult) {
	bg := context.WithoutCancel(ctx)
	for _, t := range r.Trades {
		s.publisher.Publish(port.MarketTopic(m.ID), port.StreamMessage{Type: "trade", Payload: t})
	}
	if len(r.Trades) > 0 || r.Order.Status == model.OrderOpen {
		s.publishBook(bg, m.ID)
	}

	notified := map[string]bool{}
	for _, t := range r.Trades {
		for _, u := range []string{t.MakerUserID, t.TakerUserID} {
			if notified[u] {
				continue
			}
			notified[u] = true
			notify(bg, s.notifier, port.Notification{
				UserID: u, Title: "Order filled",
				Message: fmt.Sprintf("Your order in \"%s\" was filled.", m.Question),
			})
		}
	}
}

func (s *exchangeService) publishBook(ctx context.Context, marketID int64) {
	book, err := s.OrderBook(ctx, marketID, model.OutcomeYes)
	if err != nil {
		logf("publish book of market %d: %v", marketID, err)
		return
	}
	s.publisher.Publish(port.MarketTopic(marketID), port.StreamMessage{Type: "book", Payload: book})
}

func (s *exchangeService) CancelOrder(ctx context.Context, orderID int64, userID string) (*model.Order, error) {
	peek, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if peek.UserID != userID {
		return nil, port.ErrForbidden
	}

	var order *model.Order
	var market *model.Market
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		m, err := s.markets.GetByIDForUpdate(ctx, peek.MarketID)
		if err != nil {
			return err
		}
		o, err := s.orders.GetByIDForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if o.Status != model.OrderOpen {
			return port.ErrOrderNotOpen
		}
		sess := s.session()
		if err := closeOrder(ctx, sess, o, model.OrderCanceled); err != nil {
			return err
		}
		if err := sess.flush(ctx); err != nil {
			return err
		}
		if err := s.refreshMarket(ctx, m, nil); err != nil {
			return err
		}
		order, market = o, m
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.publishBook(context.WithoutCancel(ctx), market.ID)
	return order, nil
}

func (s *exchangeService) CancelAll(ctx context.Context, marketID int64, userID string) (int, error) {
	if userID == "" {
		return 0, port.ErrInvalidUser
	}
	count := 0
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		count = 0
		m, err := s.markets.GetByIDForUpdate(ctx, marketID)
		if err != nil {
			return err
		}
		page, err := s.orders.List(ctx, port.OrderFilter{UserID: userID, MarketID: marketID, Status: model.OrderOpen, Limit: 1000})
		if err != nil {
			return err
		}
		sess := s.session()
		for i := range page.Items {
			o, err := s.orders.GetByIDForUpdate(ctx, page.Items[i].ID)
			if err != nil {
				return err
			}
			if o.Status != model.OrderOpen {
				continue
			}
			if err := closeOrder(ctx, sess, o, model.OrderCanceled); err != nil {
				return err
			}
			count++
		}
		if err := sess.flush(ctx); err != nil {
			return err
		}
		return s.refreshMarket(ctx, m, nil)
	})
	if err == nil && count > 0 {
		s.publishBook(context.WithoutCancel(ctx), marketID)
	}
	return count, err
}

func (s *exchangeService) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	return s.orders.GetByID(ctx, orderID)
}

func (s *exchangeService) ListOrders(ctx context.Context, f port.OrderFilter) (*port.Page[model.Order], error) {
	if f.UserID == "" {
		return nil, port.ErrInvalidUser
	}
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	return s.orders.List(ctx, f)
}

func (s *exchangeService) OrderBook(ctx context.Context, marketID int64, outcome model.Outcome) (*port.OrderBook, error) {
	if !outcome.Valid() {
		return nil, port.ErrInvalidOutcome
	}
	m, err := s.markets.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}
	yesBids, err := s.orders.BookLevels(ctx, marketID, model.Bid)
	if err != nil {
		return nil, err
	}
	yesAsks, err := s.orders.BookLevels(ctx, marketID, model.Ask)
	if err != nil {
		return nil, err
	}

	book := &port.OrderBook{MarketID: marketID, Outcome: outcome, Bids: yesBids, Asks: yesAsks, LastTrade: m.LastPrice}
	if outcome == model.OutcomeNo {
		book.Bids, book.Asks = flipLevels(yesAsks, m.ShareValue), flipLevels(yesBids, m.ShareValue)
		if m.LastPrice > 0 {
			book.LastTrade = m.ShareValue - m.LastPrice
		}
	}
	if book.Bids == nil {
		book.Bids = []port.BookLevel{}
	}
	if book.Asks == nil {
		book.Asks = []port.BookLevel{}
	}
	if len(book.Bids) > 0 {
		book.BestBid = book.Bids[0].YesPrice
	}
	if len(book.Asks) > 0 {
		book.BestAsk = book.Asks[0].YesPrice
	}
	if book.BestBid > 0 && book.BestAsk > 0 {
		book.Spread = book.BestAsk - book.BestBid
		book.Midpoint = float64(book.BestBid+book.BestAsk) / 2
	}
	return book, nil
}

func flipLevels(levels []port.BookLevel, shareValue int64) []port.BookLevel {
	out := make([]port.BookLevel, len(levels))
	for i, l := range levels {
		out[i] = port.BookLevel{YesPrice: shareValue - l.YesPrice, Size: l.Size}
	}
	return out
}

func (s *exchangeService) Quote(ctx context.Context, in port.QuoteInput) (*port.Quote, error) {
	if !in.Outcome.Valid() {
		return nil, port.ErrInvalidOutcome
	}
	if in.Amount <= 0 {
		return nil, port.ErrInvalidSize
	}
	book, err := s.OrderBook(ctx, in.MarketID, in.Outcome)
	if err != nil {
		return nil, err
	}

	levels := book.Asks
	if in.Side == model.Sell {
		levels = book.Bids
	}
	q := &port.Quote{}
	for _, l := range levels {
		if in.Side == model.Buy {
			n := min(l.Size, s.budgetShares(in.Amount-q.Cash-q.Fee, l.YesPrice))
			if n <= 0 {
				break
			}
			q.Shares += n
			q.Cash += n * l.YesPrice
			q.Fee += s.takerFee(n * l.YesPrice)
			q.WorstPrice = l.YesPrice
			if n < l.Size || in.Amount-q.Cash-q.Fee < l.YesPrice {
				q.Fillable = true
			}
		} else {
			n := min(l.Size, in.Amount-q.Shares)
			q.Shares += n
			q.Cash += n * l.YesPrice
			q.Fee += s.takerFee(n * l.YesPrice)
			q.WorstPrice = l.YesPrice
			if q.Shares == in.Amount {
				q.Fillable = true
				break
			}
		}
	}
	if q.Shares > 0 {
		q.AvgPrice = float64(q.Cash) / float64(q.Shares)
	}
	return q, nil
}

func (s *exchangeService) MarketTrades(ctx context.Context, marketID int64, limit, offset int) ([]model.Trade, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	return s.trades.ListByMarket(ctx, marketID, limit, max(offset, 0))
}

var ranges = map[string]struct{ window, bucket time.Duration }{
	"1h":  {time.Hour, time.Minute},
	"6h":  {6 * time.Hour, 5 * time.Minute},
	"1d":  {24 * time.Hour, 15 * time.Minute},
	"1w":  {7 * 24 * time.Hour, time.Hour},
	"1m":  {30 * 24 * time.Hour, 6 * time.Hour},
	"max": {0, 24 * time.Hour},
}

func (s *exchangeService) PriceHistory(ctx context.Context, in port.PriceHistoryInput) ([]port.Candle, error) {
	if !in.Outcome.Valid() {
		return nil, port.ErrInvalidOutcome
	}
	m, err := s.markets.GetByID(ctx, in.MarketID)
	if err != nil {
		return nil, err
	}
	r, ok := ranges[in.Range]
	if !ok {
		r = ranges["1d"]
	}
	bucket := in.Bucket
	if bucket <= 0 {
		bucket = r.bucket
	}
	var since time.Time
	if r.window > 0 {
		since = s.now().Add(-r.window)
	}
	trades, err := s.trades.ListByMarketSince(ctx, in.MarketID, since, 20000)
	if err != nil {
		return nil, err
	}
	return buildCandles(trades, in.Outcome, m.ShareValue, bucket), nil
}

func buildCandles(trades []model.Trade, outcome model.Outcome, shareValue int64, bucket time.Duration) []port.Candle {
	byStart := map[int64]*port.Candle{}
	for _, t := range trades {
		start := t.CreatedAt.Truncate(bucket)
		price := model.OwnPrice(outcome, t.YesPrice, shareValue)
		c, ok := byStart[start.Unix()]
		if !ok {
			byStart[start.Unix()] = &port.Candle{Time: start, Open: price, High: price, Low: price, Close: price, Volume: t.Size}
			continue
		}
		c.High, c.Low, c.Close = max(c.High, price), min(c.Low, price), price
		c.Volume += t.Size
	}
	out := make([]port.Candle, 0, len(byStart))
	for _, c := range byStart {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out
}
