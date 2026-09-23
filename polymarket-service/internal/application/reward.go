package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

type RewardParams struct {
	fx.In

	Markets   port.MarketRepository
	Orders    port.OrderRepository
	Positions port.PositionRepository
	Balances  port.BalanceRepository
	Referrals port.ReferralRepository
	PnL       port.PnLRepository
	Exchange  port.ExchangeAccountRepository
	Rewards   port.RewardRepository
	Tx        port.TxManager
	Notifier  port.Notifier
	Opts      Options
}

type rewardService struct {
	markets  port.MarketRepository
	orders   port.OrderRepository
	exchange port.ExchangeAccountRepository
	rewards  port.RewardRepository
	tx       port.TxManager
	notifier port.Notifier
	opts     Options
	deps     sessionDeps
}

func NewRewardService(p RewardParams) *rewardService {
	return &rewardService{
		markets: p.Markets, orders: p.Orders, exchange: p.Exchange, rewards: p.Rewards, tx: p.Tx,
		notifier: p.Notifier, opts: p.Opts.withDefaults(),
		deps: sessionDeps{
			balances: p.Balances, positions: p.Positions, referrals: p.Referrals,
			pnl: p.PnL, exchange: p.Exchange, orders: p.Orders,
		},
	}
}

func (s *rewardService) SetMarketRewards(ctx context.Context, marketID, dailyPool, maxSpread, minSize int64) (*model.Market, error) {
	var market *model.Market
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		m, err := s.markets.GetByIDForUpdate(ctx, marketID)
		if err != nil {
			return err
		}
		if dailyPool < 0 || minSize < 0 || maxSpread < 0 || maxSpread >= m.ShareValue {
			return fmt.Errorf("%w: pool and min_size must not be negative, max_spread must be below share_value", port.ErrInvalidInput)
		}
		if dailyPool > 0 && maxSpread == 0 {
			return fmt.Errorf("%w: a rewarded market needs a max_spread", port.ErrInvalidInput)
		}
		m.RewardPool, m.RewardMaxSpread, m.RewardMinSize = dailyPool, maxSpread, minSize
		market = m
		return s.markets.Save(ctx, m)
	})
	return market, err
}

func (s *rewardService) epochPool(m *model.Market) int64 {
	return m.RewardPool * int64(s.opts.RewardEpoch/time.Second) / 86_400
}

func (s *rewardService) MarketRewards(ctx context.Context, marketID int64) (*port.MarketRewards, error) {
	m, err := s.markets.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}
	return &port.MarketRewards{
		MarketID: m.ID, DailyPool: m.RewardPool, MaxSpread: m.RewardMaxSpread, MinSize: m.RewardMinSize,
		EpochSeconds: int64(s.opts.RewardEpoch / time.Second), EpochPool: s.epochPool(m),
	}, nil
}

func (s *rewardService) UserRewards(ctx context.Context, userID string, limit int, cursor string) (*port.Page[model.RewardPayout], *port.UserRewards, error) {
	if userID == "" {
		return nil, nil, port.ErrInvalidUser
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page, err := s.rewards.ListPayouts(ctx, userID, limit, cursor)
	if err != nil {
		return nil, nil, err
	}
	total, err := s.rewards.TotalPaid(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return page, &port.UserRewards{TotalPaid: total}, nil
}

func (s *rewardService) RunEpoch(ctx context.Context, now time.Time) (int, error) {
	markets, err := s.markets.ListRewarded(ctx)
	if err != nil || len(markets) == 0 {
		return 0, err
	}
	start := now.Truncate(s.opts.RewardEpoch)
	claimed, err := s.rewards.ClaimEpoch(ctx, start)
	if err != nil || !claimed {
		return 0, err
	}

	paid := 0
	for i := range markets {
		n, err := s.payMarket(ctx, &markets[i], start, now)
		if err != nil {
			logf("liquidity rewards for market %d: %v", markets[i].ID, err)
			continue
		}
		paid += n
	}
	return paid, nil
}

func (s *rewardService) payMarket(ctx context.Context, m *model.Market, epoch, now time.Time) (int, error) {
	pool := s.epochPool(m)
	if pool <= 0 || !m.Tradable(now) {
		return 0, nil
	}
	orders, err := s.orders.ListOpenByMarket(ctx, m.ID)
	if err != nil {
		return 0, err
	}
	scores := rewardScores(m, orders, now)
	var total float64
	for _, sc := range scores {
		total += sc
	}
	if total <= 0 {
		return 0, nil
	}

	var payouts []model.RewardPayout
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		payouts = nil
		revenue, err := s.exchange.Balance(ctx, model.BucketRevenue)
		if err != nil {
			return err
		}
		budget := min(pool, revenue)
		if budget <= 0 {
			return nil
		}

		users := make([]string, 0, len(scores))
		for u := range scores {
			users = append(users, u)
		}
		sort.Strings(users)

		sess := newSession(s.deps)
		var spent int64
		for _, u := range users {
			amount := int64(math.Floor(float64(budget) * scores[u] / total))
			if amount <= 0 {
				continue
			}
			bal, err := sess.balance(ctx, u)
			if err != nil {
				return err
			}
			bal.Available += amount
			sess.addPnL(u, m.ID, model.PnLReward, amount)
			payouts = append(payouts, model.RewardPayout{EpochStart: epoch, MarketID: m.ID, UserID: u, Score: scores[u], Amount: amount})
			spent += amount
		}
		if spent == 0 {
			return nil
		}
		sess.addExchange(model.BucketRevenue, "liquidity_reward", -spent, m.ID, "")
		if err := sess.flush(ctx); err != nil {
			return err
		}
		return s.rewards.CreatePayouts(ctx, payouts)
	})
	if err != nil {
		return 0, err
	}

	bg := context.WithoutCancel(ctx)
	for _, p := range payouts {
		notify(bg, s.notifier, port.Notification{
			UserID: p.UserID, Title: "Liquidity reward",
			Message: fmt.Sprintf("You earned %d for quoting \"%s\".", p.Amount, m.Question),
		})
	}
	return len(payouts), nil
}

func rewardScores(m *model.Market, orders []model.Order, now time.Time) map[string]float64 {
	var mid float64
	switch {
	case m.BestBid > 0 && m.BestAsk > 0:
		mid = float64(m.BestBid+m.BestAsk) / 2
	case m.LastPrice > 0:
		mid = float64(m.LastPrice)
	default:
		return nil
	}
	spread := float64(m.RewardMaxSpread)
	if spread <= 0 {
		return nil
	}

	type sides struct{ bid, ask float64 }
	byUser := map[string]*sides{}
	for _, o := range orders {
		if o.Remaining() < max(m.RewardMinSize, 1) || (o.ExpiresAt != nil && !now.Before(*o.ExpiresAt)) {
			continue
		}
		dist := math.Abs(float64(o.YesPrice) - mid)
		if dist > spread {
			continue
		}
		w := (spread - dist) / spread
		score := w * w * float64(o.Remaining())

		u := byUser[o.UserID]
		if u == nil {
			u = &sides{}
			byUser[o.UserID] = u
		}
		if o.BookSide == model.Bid {
			u.bid += score
		} else {
			u.ask += score
		}
	}

	extreme := mid < float64(m.ShareValue)*0.1 || mid > float64(m.ShareValue)*0.9
	out := map[string]float64{}
	for user, u := range byUser {
		switch {
		case u.bid > 0 && u.ask > 0:
			out[user] = u.bid + u.ask
		case extreme:
		default:
			out[user] = math.Max(u.bid, u.ask) / 3
		}
	}
	return out
}
