package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

type AccountParams struct {
	fx.In

	Markets   port.MarketRepository
	Trades    port.TradeRepository
	Balances  port.BalanceRepository
	Ledger    port.LedgerRepository
	Positions port.PositionRepository
	Exchange  port.ExchangeAccountRepository
	Tx        port.TxManager
	Wallets   port.WalletGateway
	Notifier  port.Notifier
	Opts      Options
}

type accountService struct {
	markets   port.MarketRepository
	trades    port.TradeRepository
	balances  port.BalanceRepository
	ledger    port.LedgerRepository
	positions port.PositionRepository
	exchange  port.ExchangeAccountRepository
	tx        port.TxManager
	wallets   port.WalletGateway
	notifier  port.Notifier
	opts      Options
	now       func() time.Time
}

func NewAccountService(p AccountParams) *accountService {
	return &accountService{
		markets: p.Markets, trades: p.Trades, balances: p.Balances, ledger: p.Ledger, positions: p.Positions,
		exchange: p.Exchange, tx: p.Tx, wallets: p.Wallets, notifier: p.Notifier, opts: p.Opts.withDefaults(), now: time.Now,
	}
}

func (s *accountService) userWallet(ctx context.Context, userID string) (*port.Wallet, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	w, err := s.wallets.GetWalletByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if w.Currency != s.opts.Currency {
		return nil, fmt.Errorf("%w: exchange trades in %s, wallet holds %s", port.ErrWalletRejected, s.opts.Currency, w.Currency)
	}
	return w, nil
}

func (s *accountService) Deposit(ctx context.Context, userID string, amount int64) (*model.Balance, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", port.ErrInvalidInput)
	}
	wallet, err := s.userWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	transferID, err := s.wallets.Transfer(ctx, port.TransferInput{
		SenderWalletID: wallet.ID, ReceiverWalletID: s.opts.TreasuryWalletID, Amount: amount,
		ReferenceID: fmt.Sprintf("polymarket:deposit:%s:%d", userID, s.now().UnixNano()),
		Description: "Polymarket deposit",
	})
	if err != nil {
		return nil, err
	}

	var balance *model.Balance
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		b, err := s.balances.GetForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		b.Available += amount
		if err := s.balances.Save(ctx, b); err != nil {
			return err
		}
		balance = b
		return s.ledger.Create(ctx, &model.LedgerEntry{UserID: userID, Type: model.LedgerDeposit, Amount: amount, TransferID: transferID})
	})
	if err != nil {
		if rerr := s.wallets.ReverseTransfer(context.WithoutCancel(ctx), transferID, "polymarket deposit failed to credit: "+err.Error()); rerr != nil {
			logf("RECONCILE: deposit transfer %s moved %d but crediting failed and reversal failed: %v", transferID, amount, rerr)
		}
		return nil, err
	}
	notify(context.WithoutCancel(ctx), s.notifier, port.Notification{
		UserID: userID, Title: "Deposit received", Message: fmt.Sprintf("%d %s was added to your trading balance.", amount, s.opts.Currency),
	})
	return balance, nil
}

func (s *accountService) Withdraw(ctx context.Context, userID string, amount int64) (*model.Balance, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", port.ErrInvalidInput)
	}
	wallet, err := s.userWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	var entry *model.LedgerEntry
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		b, err := s.balances.GetForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		if b.Available < amount {
			return port.ErrInsufficientBalance
		}
		b.Available -= amount
		if err := s.balances.Save(ctx, b); err != nil {
			return err
		}
		entry = &model.LedgerEntry{UserID: userID, Type: model.LedgerWithdraw, Amount: -amount}
		return s.ledger.Create(ctx, entry)
	})
	if err != nil {
		return nil, err
	}

	_, err = s.wallets.Transfer(ctx, port.TransferInput{
		SenderWalletID: s.opts.TreasuryWalletID, ReceiverWalletID: wallet.ID, Amount: amount,
		ReferenceID: fmt.Sprintf("polymarket:withdraw:%d", entry.ID), Description: "Polymarket withdrawal",
	})
	if err != nil {
		rerr := s.tx.WithinTx(context.WithoutCancel(ctx), func(ctx context.Context) error {
			b, err := s.balances.GetForUpdate(ctx, userID)
			if err != nil {
				return err
			}
			b.Available += amount
			if err := s.balances.Save(ctx, b); err != nil {
				return err
			}
			return s.ledger.Create(ctx, &model.LedgerEntry{UserID: userID, Type: model.LedgerWithdrawRevert, Amount: amount})
		})
		if rerr != nil {
			logf("RECONCILE: withdrawal %d of %d for %s failed at the wallet and the balance could not be restored: %v", entry.ID, amount, userID, rerr)
		}
		return nil, err
	}
	return s.balances.Get(ctx, userID)
}

func (s *accountService) Balance(ctx context.Context, userID string) (*model.Balance, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	return s.balances.Get(ctx, userID)
}

func (s *accountService) Wallet(ctx context.Context, userID string) (*port.WalletBalance, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	w, err := s.wallets.GetWalletByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &port.WalletBalance{WalletID: w.ID, Balance: w.Balance, Currency: w.Currency, Status: w.Status}, nil
}

func (s *accountService) marketsByID(ctx context.Context, ids []int64) (map[int64]model.Market, error) {
	out := map[int64]model.Market{}
	if len(ids) == 0 {
		return out, nil
	}
	markets, err := s.markets.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, m := range markets {
		out[m.ID] = m
	}
	return out, nil
}

func (s *accountService) Portfolio(ctx context.Context, userID string) (*port.Portfolio, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	balance, err := s.balances.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	open, err := s.positions.ListByUser(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	closed, err := s.positions.ListByUser(ctx, userID, true)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(open))
	for i, p := range open {
		ids[i] = p.MarketID
	}
	markets, err := s.marketsByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := &port.Portfolio{Balance: *balance, Positions: []port.PositionView{}}
	for _, p := range open {
		m := markets[p.MarketID]
		price := m.PriceOf(p.Outcome)
		view := port.PositionView{
			Position: p, Question: m.Question, EventID: m.EventID, MarketStatus: m.Status,
			AvgPrice: float64(p.CostBasis) / float64(p.Shares), CurrentPrice: price,
			Value: price * float64(p.Shares),
		}
		view.UnrealizedPnL = view.Value - float64(p.CostBasis)
		out.Positions = append(out.Positions, view)
		out.PositionsValue += view.Value
		out.UnrealizedPnL += view.UnrealizedPnL
		out.RealizedPnL += p.RealizedPnL
	}
	for _, p := range closed {
		out.RealizedPnL += p.RealizedPnL
	}
	out.TotalValue = float64(balance.Available+balance.Locked) + out.PositionsValue
	return out, nil
}

func (s *accountService) ClosedPositions(ctx context.Context, userID string) ([]model.Position, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	return s.positions.ListByUser(ctx, userID, true)
}

func (s *accountService) Activity(ctx context.Context, userID string, limit int) ([]port.Activity, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	trades, err := s.trades.ListByUser(ctx, userID, limit, 0)
	if err != nil {
		return nil, err
	}
	entries, err := s.ledger.ListByUser(ctx, userID, limit, 0)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(trades))
	for _, t := range trades {
		ids = append(ids, t.MarketID)
	}
	markets, err := s.marketsByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	items := make([]port.Activity, 0, len(trades)+len(entries))
	for _, t := range trades {
		outcome, side := t.TakerOutcome, t.TakerSide
		if t.MakerUserID == userID {
			outcome, side = t.MakerOutcome, t.MakerSide
		}
		price := model.OwnPrice(outcome, t.YesPrice, markets[t.MarketID].ShareValue)
		items = append(items, port.Activity{
			Type: "trade", MarketID: t.MarketID, Outcome: outcome, Side: side, Price: price,
			Size: t.Size, Amount: price * t.Size, CreatedAt: t.CreatedAt,
		})
	}
	for _, e := range entries {
		items = append(items, port.Activity{Type: string(e.Type), MarketID: e.MarketID, Amount: e.Amount, CreatedAt: e.CreatedAt})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *accountService) FundExchange(ctx context.Context, fromUserID string, amount int64) (*port.ExchangeSummary, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", port.ErrInvalidInput)
	}
	wallet, err := s.userWallet(ctx, fromUserID)
	if err != nil {
		return nil, err
	}
	transferID, err := s.wallets.Transfer(ctx, port.TransferInput{
		SenderWalletID: wallet.ID, ReceiverWalletID: s.opts.TreasuryWalletID, Amount: amount,
		ReferenceID: fmt.Sprintf("polymarket:funding:%s:%d", fromUserID, s.now().UnixNano()),
		Description: "Polymarket exchange funding",
	})
	if err != nil {
		return nil, err
	}
	err = s.exchange.Add(ctx, model.ExchangeEntry{Bucket: model.BucketRevenue, Kind: "funding", Amount: amount, UserID: fromUserID})
	if err != nil {
		if rerr := s.wallets.ReverseTransfer(context.WithoutCancel(ctx), transferID, "polymarket funding failed to record: "+err.Error()); rerr != nil {
			logf("RECONCILE: funding transfer %s moved %d but recording failed and reversal failed: %v", transferID, amount, rerr)
		}
		return nil, err
	}
	return s.ExchangeSummary(ctx)
}

func (s *accountService) ExchangeSummary(ctx context.Context) (*port.ExchangeSummary, error) {
	revenue, err := s.exchange.Balance(ctx, model.BucketRevenue)
	if err != nil {
		return nil, err
	}
	bonds, err := s.exchange.Balance(ctx, model.BucketBond)
	if err != nil {
		return nil, err
	}
	return &port.ExchangeSummary{Revenue: revenue, BondHeld: bonds}, nil
}
